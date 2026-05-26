package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/ports"
)

const connectionTimeout = 5 * time.Second

var _ ports.Discovery = (*Service)(nil)

// Service provides connection testing and table discovery capabilities.
type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) TestSourceConnection(ctx context.Context, cfg *ports.SourceConfig) (int64, error) {
	if cfg == nil {
		return 0, fmt.Errorf("source config is required")
	}

	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	start := time.Now()
	var err error
	switch cfg.Type {
	case constant.SourceTypePostgres.String():
		err = testPostgresConnection(ctx, postgresConfigFromSource(cfg))
	case constant.SourceTypeMySQL.String():
		err = testMySQLSourceConnection(ctx, cfg)
	default:
		err = fmt.Errorf("unsupported source type: %s", cfg.Type)
	}
	if err != nil {
		return 0, err
	}
	return time.Since(start).Milliseconds(), nil
}

func (s *Service) TestSinkConnection(ctx context.Context, cfg *ports.SinkConfig) (int64, error) {
	if cfg == nil {
		return 0, fmt.Errorf("sink config is required")
	}

	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	start := time.Now()
	var err error
	switch cfg.Type {
	case constant.SinkTypePostgres.String():
		err = testPostgresConnection(ctx, postgresConfigFromSink(cfg))
	case constant.SinkTypeMySQL.String():
		err = testMySQLSinkConnection(ctx, cfg)
	case constant.SinkTypeElasticsearch.String():
		err = testElasticsearchConnection(ctx, cfg)
	case constant.SinkTypeClickhouse.String():
		err = testClickhouseConnection(ctx, cfg)
	default:
		err = fmt.Errorf("unsupported sink type: %s", cfg.Type)
	}
	if err != nil {
		return 0, err
	}
	return time.Since(start).Milliseconds(), nil
}

func (s *Service) DiscoverSourceTables(ctx context.Context, cfg *ports.SourceConfig) ([]ports.TableInfo, error) {
	if cfg == nil {
		return nil, fmt.Errorf("source config is required")
	}

	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	switch cfg.Type {
	case constant.SourceTypePostgres.String():
		return discoverPostgresSourceTables(ctx, cfg)
	case constant.SourceTypeMySQL.String():
		return discoverMySQLSourceTables(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cfg.Type)
	}
}

func (s *Service) DiscoverSinkTables(ctx context.Context, cfg *ports.SinkConfig) ([]ports.TableInfo, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sink config is required")
	}

	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	switch cfg.Type {
	case constant.SinkTypePostgres.String():
		return discoverPostgresSinkTables(ctx, cfg)
	case constant.SinkTypeMySQL.String():
		return discoverMySQLSinkTables(ctx, cfg)
	case constant.SinkTypeClickhouse.String():
		return discoverClickhouseSinkTables(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported sink type: %s", cfg.Type)
	}
}
