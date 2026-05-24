package ports

import "github.com/foden/cdc/internal/core/domain"

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
	// ackCh receives LSN/position acknowledgements from downstream consumers.
	Start(events chan<- *domain.Event, ackCh <-chan uint64, initialOffset string) error

	// Stop gracefully shuts down the source, flushing in-flight events.
	Stop() error

	// InstanceID returns the unique identifier for this source instance.
	InstanceID() string

	// RegisterTable registers a table that a flow is interested in.
	// partitionCount determines how many partitions to hash primary keys into.
	// If the table is already registered, the refCount is incremented and
	// partitionCount is updated to the maximum of existing and new values.
	RegisterTable(schema, table string, partitionCount int)

	// UnregisterTable decrements the reference count for a table.
	// When refCount reaches zero, the table is removed from the registry.
	UnregisterTable(schema, table string)
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
