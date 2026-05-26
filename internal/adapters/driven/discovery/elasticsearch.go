package discovery

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v9"

	"github.com/foden/cdc/internal/core/ports"
)

func testElasticsearchConnection(_ context.Context, cfg *ports.SinkConfig) error {
	esCfg := elasticsearch.Config{
		Addresses: cfg.URL,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			ResponseHeaderTimeout: connectionTimeout,
		},
	}
	if cfg.APIKey != "" {
		esCfg.APIKey = cfg.APIKey
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned error: %s", res.Status())
	}
	return nil
}
