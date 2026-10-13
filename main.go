package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/commands"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/config"
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/state"
)

func loginHandler(s *state.State, cmd commands.Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("the login command expects a single argument: the username")
	}

	username := cmd.Args[0]

	if err := s.Cfg.SetUser(username); err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("User has been set to: %s\n", username)
	return nil
}

func main() {
	programState := state.State{}

	if cfg, err := config.Read(); err != nil {
		log.Fatalf("error reading config: %v\n", err)
	} else {
		programState.Cfg = &cfg
	}

	cmds := commands.Commands{
		RegisteredCommands: make(map[string]func(*state.State, commands.Command) error),
	}

	cmds.Register("login", loginHandler)

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

	fmt.Printf("%+v\n", *programState.Cfg)
}
