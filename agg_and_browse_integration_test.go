package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/config"
)

func TestAggAndBrowseIntegration(t *testing.T) {
	// // Read the same database URL your normal CLI uses.
	// cfg, err := config.Read()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	dbURL := os.Getenv("GATOR_TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("set GATOR_TEST_DATABASE_URL before running this test")
	}

	// Build the CLI once, then run it for each test step.
	binary := filepath.Join(t.TempDir(), "gator")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}

	// Give the CLI a private config file. It uses your existing database,
	// but won't change the current user in your real config file.
	home := t.TempDir()
	testConfig, err := json.Marshal(config.Config{
		DBUrl: dbURL,
		// cfg.DBUrl
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".gatorconfig.json"), testConfig, 0o600); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(), "HOME="+home)

	run := func(args ...string) string {
		t.Helper()

		cmd := exec.Command(binary, args...)
		cmd.Env = env
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gator %v failed: %v\n%s", args, err, output)
		}
		return string(output)
	}

	// Serve predictable RSS locally instead of contacting public websites.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")

		switch r.URL.Path {
		case "/alice":
			fmt.Fprint(w, `<rss><channel><title>Alice</title>
		<item><title>Alice's first article</title><link>https://example.test/alice-1</link><pubDate>Mon, 06 Sep 2021 12:00:00 GMT</pubDate></item>
		<item><title>Alice's second article</title><link>https://example.test/alice-2</link><pubDate>Tue, 07 Sep 2021 12:00:00 GMT</pubDate></item>
		<item><title>Alice's third article</title><link>https://example.test/alice-3</link><pubDate>Wed, 08 Sep 2021 12:00:00 GMT</pubDate></item>
	</channel></rss>`)
		case "/bob":
			fmt.Fprint(w, `<rss><channel><title>Bob</title>
		<item><title>Bob's first article</title><link>https://example.test/bob-1</link><pubDate>Mon, 06 Sep 2021 14:30:00 GMT</pubDate></item>
		<item><title>Bob's second article</title><link>https://example.test/bob-2</link><pubDate>Tue, 07 Sep 2021 14:30:00 GMT</pubDate></item>
		<item><title>Bob's third article</title><link>https://example.test/bob-3</link><pubDate>Wed, 08 Sep 2021 14:30:00 GMT</pubDate></item>
	</channel></rss>`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	run("reset")
	run("register", "alice")
	run("addfeed", "Alice's feed", server.URL+"/alice")
	run("register", "bob")
	run("addfeed", "Bob's feed", server.URL+"/bob")

	// agg should run continuously. Stop it after three seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, "agg", "200ms")
	cmd.Env = env
	var output bytes.Buffer
	liveOutput := io.MultiWriter(os.Stdout, &output)
	cmd.Stdout = liveOutput
	cmd.Stderr = liveOutput

	err = cmd.Run()
	if err != nil && ctx.Err() == nil {
		t.Fatalf("agg failed: %v\n%s", err, output.String())
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, tc := range []struct {
		url   string
		title string
	}{
		{"https://example.test/alice-1", "Alice's first article"},
		{"https://example.test/alice-2", "Alice's second article"},
		{"https://example.test/alice-3", "Alice's third article"},
		{"https://example.test/bob-1", "Bob's first article"},
		{"https://example.test/bob-2", "Bob's second article"},
		{"https://example.test/bob-3", "Bob's third article"},
	} {
		var title string
		var count int
		err := db.QueryRow(
			"SELECT count(*), COALESCE(MAX(title), '') FROM posts WHERE url = $1",
			tc.url,
		).Scan(&count, &title)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 || title != tc.title {
			t.Errorf(
				"post %s: got count=%d title=%q; want count=1 title=%q\nagg output:\n%s",
				tc.url, count, title, tc.title, output.String(),
			)
		}
	}

	// Bob initially follows only the feed he added. His browse output
	// should not include posts from Alice's feed.
	bobPosts := run("browse", "10")
	if !strings.Contains(bobPosts, "Bob's first article") {
		t.Errorf("Bob cannot browse his followed feed:\n%s", bobPosts)
	}
	if strings.Contains(bobPosts, "Alice's first article") {
		t.Errorf("Bob sees Alice's unfollowed feed:\n%s", bobPosts)
	}

	// Following Alice's feed should make its posts visible to Bob, even
	// though Alice—not Bob—created that feed.
	run("follow", server.URL+"/alice")
	bobPosts = run("browse", "10")
	for _, title := range []string{
		"Alice's first article",
		"Alice's second article",
		"Alice's third article",
		"Bob's first article",
		"Bob's second article",
		"Bob's third article",
	} {
		if !strings.Contains(bobPosts, title) {
			t.Errorf("Bob's browse output missing %q:\n%s", title, bobPosts)
		}
	}

	// Without an argument, browse should default to two posts.
	defaultBrowse := run("browse")
	if got := strings.Count(defaultBrowse, "https://example.test/"); got != 2 {
		t.Errorf("default browse showed %d post URLs; want 2:\n%s", got, defaultBrowse)
	}

	// Switching users should change which followed feeds are visible.
	run("login", "alice")
	alicePosts := run("browse", "10")
	if !strings.Contains(alicePosts, "Alice's first article") {
		t.Errorf("Alice cannot browse her followed feed:\n%s", alicePosts)
	}
	if strings.Contains(alicePosts, "Bob's first article") {
		t.Errorf("Alice sees Bob's unfollowed feed:\n%s", alicePosts)
	}

}
