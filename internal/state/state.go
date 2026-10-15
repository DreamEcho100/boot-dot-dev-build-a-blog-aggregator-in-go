package state

import (
	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/config"

	"github.com/DreamEcho100/boot-dot-dev-build-a-blog-aggregator-in-go/internal/database"
)

type State struct {
	Cfg *config.Config
	DB  *database.Queries
}
