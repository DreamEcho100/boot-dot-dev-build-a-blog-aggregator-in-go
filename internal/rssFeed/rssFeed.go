package rssFeed

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/database"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/state"
	"github.com/lib/pq"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func parsePublishedAt(value string) sql.NullTime {
	value = strings.TrimSpace(value)

	for _, layout := range []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
	} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}

	// Missing or unrecognized date: store SQL NULL.
	return sql.NullTime{}
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result RSSFeed
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	result.Channel.Title = html.UnescapeString(result.Channel.Title)
	result.Channel.Description = html.UnescapeString(result.Channel.Description)

	for i := range result.Channel.Item {
		item := &result.Channel.Item[i]
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return &result, nil
}

func ScrapeFeeds(ctx context.Context, s *state.State, time_between_reqs string) error {
	duration, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		feed, err := s.DB.GetNextFeedToFetch(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("No more feeds to fetch!")
			return nil
		} else if err != nil {
			return err
		}

		log.Printf("fetching feed %q (%s)", feed.Name, feed.Url)
		rssFeed, err := FetchFeed(ctx, feed.Url)
		if err != nil {
			return err
		}

		err = s.DB.MarkFeedFetched(ctx, feed.ID)
		if err != nil {
			return err
		}

		for _, item := range rssFeed.Channel.Item {
			// The schema requires a unique, nonempty URL. An item without one
			// cannot be inserted as a useful post.
			if strings.TrimSpace(item.Link) == "" {
				log.Printf("skipping item %q from %s: no link", item.Title, feed.Url)
				continue
			}

			_, err = s.DB.CreatePost(ctx, database.CreatePostParams{
				Title:       item.Title,
				Url:         item.Link,
				Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
				FeedID:      feed.ID,
				PublishedAt: parsePublishedAt(item.PubDate),
			})
			if err != nil {
				var pqErr *pq.Error
				if errors.As(err, &pqErr) && pqErr.Code == "23505" {
					// Already stored this URL; don't log every repeat scrape.
					continue
				}

				log.Printf("saving post %q from %s: %v", item.Link, feed.Url, err)
			} else {
				log.Printf("saved post %q from %s", item.Title, feed.Url)
			}

		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

	}
}
