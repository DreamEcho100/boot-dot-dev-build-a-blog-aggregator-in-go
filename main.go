package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/commands"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/config"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/database"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/rssFeed"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/state"
)

func loginHandler(s *state.State, cmd commands.Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("the login command expects a single argument: the username")
	}

	username := cmd.Args[0]

	res, err := s.DB.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("error in query: %v", err)
	}

	if err := s.Cfg.SetUser(username); err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("User has been set to: %+v\n", res)
	return nil
}

func registerHandler(s *state.State, cmd commands.Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("the register command expects a single argument: the username")
	}

	username := cmd.Args[0]
	CreatedAt := time.Now().UTC()
	res, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      username,
		CreatedAt: CreatedAt,
		UpdatedAt: CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("error in query: %v", err)
	}

	if err := s.Cfg.SetUser(username); err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("User has been registered: %+v\n", res)
	return nil
}

func resetHandler(s *state.State, cmd commands.Command) error {
	err := s.DB.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error in query: %v", err)
	}

	err = s.DB.ResetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error in query: %v", err)
	}

	fmt.Println("Successful reset operation!")
	return nil
}

func getUsersHandler(s *state.State, cmd commands.Command) error {
	res, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error in query: %v", err)
	}

	for _, user := range res {
		fmt.Printf("* %s", user.Name)
		if user.Name == s.Cfg.CurrentUsername {
			fmt.Println(" (current)")
		} else {
			fmt.Println()
		}
	}

	return nil
}

func aggHandler(s *state.State, cmd commands.Command) error {
	if len(cmd.Args) != 1 {
		return errors.New("the agg command expects a single argument: the time between reqs _(1s, 1m, 1h, 1d, 1m1s, 1h3m2s, etc.)_")
	}

	time_between_reqs := cmd.Args[0]

	err := rssFeed.ScrapeFeeds(context.Background(), s, time_between_reqs)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}

	return nil
}

func addFeedHandler(s *state.State, cmd commands.Command, currentUser database.User) error {
	if len(cmd.Args) < 2 {
		return errors.New("the addfeed command expects at least two arguments: the name and the url feed on this order")
	}

	feedName := cmd.Args[0]
	urlInput := cmd.Args[1]
	_, err := url.Parse(urlInput)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	result, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:     uuid.New(),
		Name:   feedName,
		Url:    urlInput,
		UserID: currentUser.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed: %w", err)
	}

	_, err = s.DB.CreateFeedFollowByFeedUrl(context.Background(), database.CreateFeedFollowByFeedUrlParams{
		UserID: currentUser.ID,
		Url:    urlInput,
	})
	if err != nil {
		return fmt.Errorf("error following new feed: %w", err)
	}

	fmt.Printf("Feed created successfully! ID: %s\n", result.ID)
	return nil
}

func getFeedsHandler(s *state.State, cmd commands.Command) error {
	result, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds %w", err)
	}

	for _, feed := range result {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(feed.UserName)
	}

	return nil
}

func followHandler(s *state.State, cmd commands.Command, currentUser database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("the command expects at least one argument: the url of the feed!")
	}

	urlInput := cmd.Args[0]
	_, err := url.Parse(urlInput)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	result, err := s.DB.CreateFeedFollowByFeedUrl(context.Background(), database.CreateFeedFollowByFeedUrlParams{
		Url:    urlInput,
		UserID: currentUser.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}

	fmt.Println(result.FeedName)
	fmt.Println(result.UserName)
	return nil
}

func followingHandler(s *state.State, cmd commands.Command, currentUser database.User) error {
	result, err := s.DB.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("error getting current user follows: %w", err)
	}

	for _, item := range result {
		fmt.Println(item.Name)
		// fmt.Println(item.AuthorName)
	}

	return nil
}

func unfollowHandler(s *state.State, cmd commands.Command, currentUser database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("the command expects at least one argument: the url of the feed!")
	}

	urlInput := cmd.Args[0]
	_, err := url.Parse(urlInput)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	err = s.DB.DeleteFeedFollowByUrl(context.Background(), database.DeleteFeedFollowByUrlParams{
		Url:    urlInput,
		UserID: currentUser.ID,
	})
	if err != nil {
		return fmt.Errorf("error in the delete feed follow query %w", err)
	}

	return nil
}

func browseHandler(s *state.State, cmd commands.Command, currentUser database.User) error {
	if len(cmd.Args) > 1 {
		return errors.New("usage: browse [limit]!")
	}

	limit := 2
	if len(cmd.Args) == 1 {
		parsed, err := strconv.Atoi(cmd.Args[0])
		if err != nil || parsed <= 0 {
			return errors.New("browse limit must be a positive integer")
		}
		limit = parsed
	}

	posts, err := s.DB.GetPostsByUserId(context.Background(), database.GetPostsByUserIdParams{
		UserID: currentUser.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("getting posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found. Try running agg to fetch your feeds.")
		return nil
	}

	for _, post := range posts {
		fmt.Println(post.Title)
		fmt.Println(post.Url)
		if post.Description.Valid && post.Description.String != "" {
			fmt.Println(post.Description.String)
		}
		fmt.Println()
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state.State, cmd commands.Command, currentUser database.User) error) func(*state.State, commands.Command) error {
	return func(s *state.State, cmd commands.Command) error {
		currentUser, err := s.DB.GetUser(context.Background(), s.Cfg.CurrentUsername)
		if err != nil {
			return fmt.Errorf("error getting current user: %w", err)
		}

		return handler(s, cmd, currentUser)
	}
}

func main() {
	programState := state.State{}

	if cfg, err := config.Read(); err != nil {
		log.Fatalf("error reading config: %v\n", err)
	} else {
		programState.Cfg = &cfg
	}

	db, err := sql.Open("postgres", programState.Cfg.DBUrl)
	if err != nil {
		log.Fatalf("error opening the DB connection: %v", err)
	}
	dbQueries := database.New(db)
	programState.DB = dbQueries

	cmds := commands.Commands{
		RegisteredCommands: make(map[string]func(*state.State, commands.Command) error),
	}

	cmds.Register("login", loginHandler)
	cmds.Register("register", registerHandler)
	cmds.Register("reset", resetHandler)
	cmds.Register("users", getUsersHandler)
	cmds.Register("agg", aggHandler)
	cmds.Register("addfeed", middlewareLoggedIn(addFeedHandler))
	cmds.Register("feeds", getFeedsHandler)
	cmds.Register("follow", middlewareLoggedIn(followHandler))
	cmds.Register("following", middlewareLoggedIn(followingHandler))
	cmds.Register("unfollow", middlewareLoggedIn(unfollowHandler))
	cmds.Register("browse", middlewareLoggedIn(browseHandler))

	if len(os.Args) < 2 {
		log.Fatalln("Usage: gator <command> [args...]")
	}

	cmd := commands.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	if err := cmds.Run(&programState, cmd); err != nil {
		log.Fatal(err)
	}
}
