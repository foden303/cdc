package ports

import (
	"context"

	"github.com/foden/cdc/internal/core/domain"
)

type SourceAck struct {
	LSN    uint64
	Offset string
}

type SourceTableRef struct {
	Schema string
	Table  string
}

type SourceTableSyncer interface {
	SyncSourceTables(ctx context.Context, tables []SourceTableRef) error
}

// Source defines a connection to a database that emits change events.
// Implementations are responsible for:
//   - Connecting to the upstream database (Postgres WAL, MySQL binlog, etc.)
//   - Converting raw replication events into domain.Event
//   - Hashing primary keys to determine partition IDs for NATS subjects
//   - Handling reconnection with exponential backoff
//
// Sources do NOT own flow-level concerns (filtering, batching, sink routing).
// Those are managed by the FlowManager.
type Source interface {
	// Start begins streaming CDC events into the provided channel.
	// initialOffset allows resuming from a previously checkpointed position.
	// ackCh receives source-position acknowledgements after durable publish.
	Start(events chan<- *domain.Event, ackCh <-chan SourceAck, initialOffset string) error

	// Stop gracefully shuts down the source, flushing in-flight events.
	Stop() error

	// InstanceID returns the unique identifier for this source instance.
	InstanceID() string
}

// SourceConfig holds connection-level fields for a CDC source.
// This struct is shared across all source types (Postgres, MySQL, MariaDB).
// It lives in the interfaces package because multiple packages (source, discovery,
// server, flow) need to reference it without creating circular imports.
type SourceConfig struct {
	InstanceID string `json:"instance_id"`
	Name       string `json:"name"`
	Type       string `json:"type"` // Use constant.SourceType* for valid values
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Database   string `json:"database"`
}
