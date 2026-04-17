package models

import (
	"encoding/json"
)

// Event represents a CDC event envelope for routing and transport.
type Event struct {
	// Metadata for routing and management
	Topic      string `json:"-"`
	Subject    string `json:"-"`
	InstanceID string `json:"instance_id"`
	Schema     string `json:"schema"`
	Table      string `json:"table"`
	Op         string `json:"op"`
	Offset     string `json:"offset"`
	LSN        uint64 `json:"lsn"`
	// Raw Debezium Payload
	Data []byte `json:"-"`
	// Meta for partitioning
	Partition int `json:"partition"`
}

// DebeziumPayload represents the industry-standard CDC format.
type DebeziumPayload struct {
	Op          string          `json:"op"` // "c", "u", "d", "r"
	Before      json.RawMessage `json:"before,omitempty"`
	After       json.RawMessage `json:"after,omitempty"`
	Source      SourceMetadata  `json:"source"`
	TimestampMS int64           `json:"ts_ms"`
}

// SourceMetadata contains origin information for the event.
type SourceMetadata struct {
	Version   string `json:"version"`
	Connector string `json:"connector"`
	Name      string `json:"name"`
	TsMs      int64  `json:"ts_ms"`
	Snapshot  string `json:"snapshot"`
	DB        string `json:"db"`
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	LSN       uint64 `json:"lsn"`
	TxId      int64  `json:"txId,omitempty"`
}

// NewEvent creates a new CDC event envelope.
func NewEvent(topic, subject, instanceID, schema, table, op string, lsn uint64, offset string, data []byte, partition int) *Event {
	return &Event{
		Topic:      topic,
		Subject:    subject,
		InstanceID: instanceID,
		Schema:     schema,
		Table:      table,
		Op:         op,
		LSN:        lsn,
		Offset:     offset,
		Data:       data,
		Partition:  partition,
	}
}

// Reset clears the event fields for reuse.
func (e *Event) Reset() {
	e.Topic = ""
	e.Subject = ""
	e.InstanceID = ""
	e.Schema = ""
	e.Table = ""
	e.Op = ""
	e.Offset = ""
	e.LSN = 0
	e.Data = nil
	e.Partition = 0
}

// ComponentStats represents success/failure metrics for a source or sink.
type ComponentStats struct {
	SuccessCount uint64           `json:"success_count"`
	FailureCount uint64           `json:"failure_count"`
	LastError    string           `json:"last_error"`
	PartitionLag map[int32]uint64 `json:"partition_lag"`
}

type MessageStatus int

const (
	MessageStatusSent MessageStatus = iota
	MessageStatusUnsent
	MessageStatusAll
)

type Message struct {
	Sequence  uint64            `json:"sequence"`
	Timestamp int64             `json:"timestamp"`
	Subject   string            `json:"subject"`
	Data      []byte            `json:"data"`
	Headers   map[string]string `json:"headers"`
}
