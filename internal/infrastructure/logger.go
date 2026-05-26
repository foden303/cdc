package infrastructure

import (
	"github.com/foden/cdc/config"
	"github.com/foden/cdc/pkg/logger"
)

func SetupLogger(cfg config.LogConfig) {
	logger.Init(cfg)
}
