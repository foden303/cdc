package global

import (
	"log/slog"

	"github.com/foden/cdc/config"
)

var (
	Config *config.Config
	Logger *slog.Logger
)
