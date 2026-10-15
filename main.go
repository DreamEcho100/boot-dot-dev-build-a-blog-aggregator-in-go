package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/commands"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/config"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/database"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/state"
	"github.com/google/uuid"
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
	err := s.DB.Reset(context.Background())
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
		if user.Name == s.Cfg.CurrentUserame {
			fmt.Println(" (current)")
		} else {
			fmt.Println()
		}
	}

	return nil
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
