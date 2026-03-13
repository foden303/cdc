package sink_consumer

import (
	"time"

	"github.com/foden/cdc/pkg/config"
)

// CreateConsumerConfig defines the configuration for creating a new sink consumer
type CreateConsumerConfig struct {
	// Name is a human-readable identifier for the consumer
	Name string `json:"name"`

	// TopicName is the queue topic to consume from
	TopicName string `json:"topic_name"`

	// Partitions is an optional list of partition IDs to subscribe to
	// If empty, consumer subscribes to all partitions
	Partitions []int `json:"partitions,omitempty"`

	// Sinks defines the destination sinks to write to
	Sinks []SinkSpec `json:"sinks"`

	// BatchSize is the number of messages to accumulate before flushing
	// Default: 500
	BatchSize int `json:"batch_size,omitempty"`

	// FlushIntervalMs is the time (in milliseconds) to wait before flushing
	// Default: 1000 (1 second)
	FlushIntervalMs int `json:"flush_interval_ms,omitempty"`
}

// SinkSpec defines a single sink destination
type SinkSpec struct {
	// Type is the sink type (e.g., "elasticsearch", "http", "clickhouse")
	Type string `json:"type"`

	// Config is the sink-specific configuration
	Config *config.SinkConfig `json:"config"`
}

// ConsumerStats provides runtime statistics for a sink consumer
type ConsumerStats struct {
	// ID is the consumer identifier
	ID string `json:"id"`

	// Name is the human-readable name
	Name string `json:"name"`

	// TopicName is the topic being consumed
	TopicName string `json:"topic_name"`

	// Status is the current state: "running", "stopped", "paused", "error"
	Status string `json:"status"`

	// TotalProcessed is the total number of messages successfully processed
	TotalProcessed uint64 `json:"total_processed"`

	// TotalErrors is the total number of messages that failed processing
	TotalErrors uint64 `json:"total_errors"`

	// LastOffset maps partition_id -> last committed offset
	LastOffset map[int]uint64 `json:"last_offset"`

	// LastUpdated is the timestamp of the last stats update
	LastUpdated time.Time `json:"last_updated"`

	// UpSince is when the consumer started
	UpSince *time.Time `json:"up_since,omitempty"`

	// Sinks lists the sink destinations being written to
	Sinks []string `json:"sinks"`
}

// ConsumerMetrics tracks internal metrics for a consumer
type ConsumerMetrics struct {
	TotalProcessed uint64
	TotalErrors    uint64
	LastUpdate     time.Time
	StartTime      time.Time
}

// UpdateConsumerConfig represents partial updates to a consumer
type UpdateConsumerConfig struct {
	// Status change: "running" or "paused"
	Status string `json:"status,omitempty"`

	// Optional: new batch size
	BatchSize *int `json:"batch_size,omitempty"`

	// Optional: new flush interval
	FlushIntervalMs *int32 `json:"flush_interval_ms,omitempty"`
}

// ConsumerListResponse wraps a list of consumer stats
type ConsumerListResponse struct {
	Consumers []*ConsumerStats `json:"consumers"`
	Total     int              `json:"total"`
}

// CreateConsumerResponse is returned after creating a consumer
type CreateConsumerResponse struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Stats ConsumerStats `json:"stats"`
}

// ErrorResponse represents API error in standard format
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
