package commands

import (
	"fmt"

	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/state"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	RegisteredCommands map[string]func(*state.State, Command) error
}

func (cmds *Commands) Register(name string, cb func(*state.State, Command) error) {
	cmds.RegisteredCommands[name] = cb
}

func (cmds *Commands) Run(s *state.State, cmd Command) error {
	handler, ok := cmds.RegisteredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("unsknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}
