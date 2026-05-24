package infrastructure

import (
	"github.com/foden/cdc/config"
	"github.com/foden/cdc/global"
)

func LoadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	global.Config = cfg
	return cfg, nil
}
