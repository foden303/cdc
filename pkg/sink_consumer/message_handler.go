package sink_consumer

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/foden/cdc/pkg/models"
	"github.com/foden/cdc/pkg/queue"
)

// MessageHandler deserializes queue messages into CDC events
type MessageHandler interface {
	Handle(msg *queue.MessageView) (*models.Event, error)
}

// JSONMessageHandler deserializes JSON-encoded CDC events
type JSONMessageHandler struct{}

// NewJSONMessageHandler creates a new JSON message handler
func NewJSONMessageHandler() *JSONMessageHandler {
	return &JSONMessageHandler{}
}

// Handle deserializes a queue message into a CDC event
// Expected format: queue.MessageView.Value contains JSON-encoded models.Event
func (h *JSONMessageHandler) Handle(msg *queue.MessageView) (*models.Event, error) {
	if len(msg.Value) == 0 {
		slog.Warn("empty message value", "offset", msg.Offset, "key", string(msg.Key))
		return nil, fmt.Errorf("empty message value at offset %d", msg.Offset)
	}

	event := &models.Event{}
	if err := json.Unmarshal(msg.Value, event); err != nil {
		slog.Error("failed to unmarshal event",
			"offset", msg.Offset,
			"err", err,
			"value_sample", string(msg.Value[:min(50, len(msg.Value))]))
		return nil, fmt.Errorf("failed to unmarshal event at offset %d: %w", msg.Offset, err)
	}

	// Validate basic event structure
	if event.Table == "" {
		slog.Warn("event missing table", "offset", msg.Offset)
		return nil, fmt.Errorf("event at offset %d missing table field", msg.Offset)
	}

	return event, nil
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
