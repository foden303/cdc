package ports

import "context"

// FlowStatus represents the current state of a flow.
type FlowStatus string

const (
	FlowStatusRunning FlowStatus = "running"
	FlowStatusPaused  FlowStatus = "paused"
	FlowStatusError   FlowStatus = "error"
)

// FlowConfig holds the complete configuration for a flow unit.
type FlowConfig struct {
	FlowID         string          `json:"flow_id"`
	Name           string          `json:"name"`
	SourceID       string          `json:"source_id"`
	SinkID         string          `json:"sink_id"`
	SourceTable    string          `json:"source_table"`
	SinkTable      string          `json:"sink_table"`
	Status         FlowStatus      `json:"status"`
	CreatedAt      int64           `json:"created_at"`
	UpdatedAt      int64           `json:"updated_at"`
	Options        *FlowOptions    `json:"options,omitempty"`
	ColumnMappings []ColumnMapping `json:"column_mappings,omitempty"`
}

// FlowOptions holds optional per-flow tuning parameters.
type FlowOptions struct {
	BatchSize        int32  `json:"batch_size,omitempty"`
	FlushIntervalMs  int32  `json:"flush_interval_ms,omitempty"`
	FilterExpression string `json:"filter_expression,omitempty"`
	PoolSize         int    `json:"pool_size,omitempty"`
	PartitionCount   int    `json:"partition_count,omitempty"` // default 4 if 0
}

// ColumnMapping defines how a source column maps to a sink column.
type ColumnMapping struct {
	SourceColumn string `json:"source_column"`
	SinkColumn   string `json:"sink_column"`
	SourceType   string `json:"source_type"`
	SinkType     string `json:"sink_type"`
	Enabled      bool   `json:"enabled"`
}

// FlowStats holds runtime statistics for a flow.
type FlowStats struct {
	EventsPerSecond      float64 `json:"events_per_second"`
	ReplicationLagMs     int64   `json:"replication_lag_ms"`
	LastSyncedAt         int64   `json:"last_synced_at"`
	TotalEventsProcessed uint64  `json:"total_events_processed"`
}

// FlowManager defines the lifecycle operations for flows.
type FlowManager interface {
	CreateFlow(ctx context.Context, cfg *FlowConfig) (*FlowConfig, error)
	GetFlow(ctx context.Context, flowID string) (*FlowConfig, error)
	ListFlows(ctx context.Context) ([]*FlowConfig, error)
	UpdateFlow(ctx context.Context, cfg *FlowConfig) (*FlowConfig, error)
	DeleteFlow(ctx context.Context, flowID string) error
	PauseFlow(ctx context.Context, flowID string) (*FlowConfig, error)
	ResumeFlow(ctx context.Context, flowID string) (*FlowConfig, error)
	GetFlowStats(ctx context.Context, flowID string) (*FlowStats, error)
	RestoreFlows(ctx context.Context) error
	Stop()
}
