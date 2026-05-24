package infrastructure

import (
	"log/slog"

	"github.com/foden/cdc/config"
	"github.com/foden/cdc/global"
	"github.com/foden/cdc/pkg/logger"
)

func SetupLogger(cfg config.LogConfig) {
	logger.Init(cfg)
	global.Logger = logger.L()
}

func log() *slog.Logger {
	if global.Logger != nil {
		return global.Logger
	}
	return slog.Default()
}
