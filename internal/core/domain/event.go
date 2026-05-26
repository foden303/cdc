package domain

import (
	"encoding/json"

	"github.com/foden/cdc/internal/core/constant"
)

// Event represents a CDC event envelope for routing and transport.
type Event struct {
	// Metadata for routing and management
	Topic       string      `json:"-"`
	Subject     string      `json:"-"`
	InstanceID  string      `json:"instance_id"`
	Schema      string      `json:"schema"`
	Table       string      `json:"table"`
	Op          constant.Op `json:"op"`
	Offset      string      `json:"offset"`
	LSN         uint64      `json:"lsn"`
	TimestampMS int64       `json:"ts_ms"`
	// Raw Debezium Payload
	Data []byte `json:"-"`
	// Optional explicit message identifier used by NATS deduplication.
	MessageID string `json:"-"`
	// Meta for partitioning
	Partition int `json:"partition"`
}

// DebeziumPayload represents the industry-standard CDC format.
type DebeziumPayload struct {
	Op          constant.Op     `json:"op"` // "c", "u", "d", "r"
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
func NewEvent(topic, subject, instanceID, schema, table string, op constant.Op, lsn uint64, offset string, data []byte, partition int) *Event {
	return &Event{
		Topic:       topic,
		Subject:     subject,
		InstanceID:  instanceID,
		Schema:      schema,
		Table:       table,
		Op:          op,
		LSN:         lsn,
		TimestampMS: 0,
		Offset:      offset,
		Data:        data,
		Partition:   partition,
	}
}

// Reset clears the event fields for reuse via sync.Pool.
func (e *Event) Reset() {
	e.Topic = ""
	e.Subject = ""
	e.InstanceID = ""
	e.Schema = ""
	e.Table = ""
	e.Op = ""
	e.Offset = ""
	e.LSN = 0
	e.TimestampMS = 0
	e.Data = nil
	e.MessageID = ""
	e.Partition = 0
}

// DeepClone creates an independent copy of the Event with no shared backing arrays.
// This is used before passing events to sinks to prevent use-after-free when the
// original is returned to the sync.Pool.
func (e *Event) DeepClone() *Event {
	clone := &Event{
		Topic:       e.Topic,
		Subject:     e.Subject,
		InstanceID:  e.InstanceID,
		Schema:      e.Schema,
		Table:       e.Table,
		Op:          e.Op,
		Offset:      e.Offset,
		LSN:         e.LSN,
		TimestampMS: e.TimestampMS,
		MessageID:   e.MessageID,
		Partition:   e.Partition,
	}
	if e.Data != nil {
		clone.Data = make([]byte, len(e.Data))
		copy(clone.Data, e.Data)
	}
	return clone
}

type MessageStatus int

const (
	MessageStatusSent MessageStatus = iota
	MessageStatusUnsent
	MessageStatusAll
)
