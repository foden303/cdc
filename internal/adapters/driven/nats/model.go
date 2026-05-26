package nats

import "github.com/foden/cdc/internal/core/ports"

// MessageItem represents a single message from the stream.
type MessageItem = ports.NATSMessageItem

// ConsumerSummary represents a flow consumer on the CDC event stream.
type ConsumerSummary = ports.NATSConsumerSummary
