package infrastructure

import (
	"log/slog"

	drivenregistry "github.com/foden/cdc/internal/adapters/driven/registry"
	"github.com/foden/cdc/internal/core/ports"

	// Connector self-registration via init() functions.
	_ "github.com/foden/cdc/internal/adapters/driven/connector/sink/clickhouse"
	_ "github.com/foden/cdc/internal/adapters/driven/connector/sink/elasticsearch"
	_ "github.com/foden/cdc/internal/adapters/driven/connector/sink/mysql"
	_ "github.com/foden/cdc/internal/adapters/driven/connector/sink/postgres"
	_ "github.com/foden/cdc/internal/adapters/driven/connector/source/mysql"
	_ "github.com/foden/cdc/internal/adapters/driven/connector/source/postgres"
)

func SetupRegistry() ports.Registry {
	reg := drivenregistry.Default()
	slog.Info("connectors registered",
		"sources", reg.SourceNames(),
		"sinks", reg.SinkNames(),
	)
	return reg
}
