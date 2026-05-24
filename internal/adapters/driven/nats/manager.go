package nats

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// CreateStream uses retention days from config
func (c *Client) CreateStream(ctx context.Context, subjects []string) error {
	maxAge := time.Duration(c.cfg.RetentionDays) * 24 * time.Hour
	if _, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:       c.streamName,
		Subjects:   subjects,
		MaxAge:     maxAge,
		Duplicates: 2 * time.Minute,
		Storage:    jetstream.FileStorage,
	}); err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	return nil
}

// CreateOrUpdateConsumer creates or updates a pull consumer for the stream.
// MaxAckPending controls ordering: set low for strict ordering, higher for throughput.
func (c *Client) CreateOrUpdateConsumer(ctx context.Context, name string, filter []string) (jetstream.Consumer, error) {
	return c.js.CreateOrUpdateConsumer(ctx, c.streamName, jetstream.ConsumerConfig{
		Durable:        name,
		FilterSubjects: filter,
		AckPolicy:      jetstream.AckExplicitPolicy,
		MaxAckPending:  c.cfg.MaxAckPending,
		AckWait:        time.Duration(c.cfg.AckWaitMs) * time.Millisecond,
		MaxDeliver:     c.cfg.MaxDeliver,
		ReplayPolicy:   jetstream.ReplayInstantPolicy,
	})
}

// CreateDLQStream creates a dedicated Dead Letter Queue stream for failed messages.
// It defaults to a 30-day retention unless specified otherwise.
func (c *Client) CreateDLQStream(ctx context.Context) error {
	retention := 30 * 24 * time.Hour
	if c.cfg.RetentionDays > 0 {
		retention = time.Duration(c.cfg.RetentionDays) * 24 * time.Hour
	}

	if _, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      c.dlqStreamName(),
		Subjects:  []string{"dlq.>"},
		MaxAge:    retention,
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
	}); err != nil {
		return fmt.Errorf("failed to create DLQ stream %s: %w", c.dlqStreamName(), err)
	}
	slog.Info("DLQ stream initialized", "stream", c.dlqStreamName(), "retention", retention)
	return nil
}

// GetStreamInfo fetches detailed metadata about the primary JetStream stream.
func (c *Client) GetStreamInfo(ctx context.Context) (*jetstream.StreamInfo, error) {
	stream, err := c.js.Stream(ctx, c.streamName)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to stream %s: %w", c.streamName, err)
	}
	return stream.Info(ctx)
}

// DeleteConsumer removes a durable consumer by name. Returns nil if the consumer does not exist.
func (c *Client) DeleteConsumer(ctx context.Context, name string) error {
	err := c.js.DeleteConsumer(ctx, c.streamName, name)
	if err != nil {
		// Treat "consumer not found" as a no-op
		if err.Error() == "consumer not found" {
			return nil
		}
		return fmt.Errorf("failed to delete consumer %s: %w", name, err)
	}
	slog.Info("NATS consumer deleted", "consumer", name, "stream", c.streamName)
	return nil
}

// GetConsumer binds to an existing durable consumer by name.
func (c *Client) GetConsumer(ctx context.Context, name string) (jetstream.Consumer, error) {
	return c.js.Consumer(ctx, c.streamName, name)
}

func (c *Client) ListFlowConsumerNames(ctx context.Context) ([]string, error) {
	stream, err := c.js.Stream(ctx, c.streamName)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to stream %s: %w", c.streamName, err)
	}

	lister := stream.ConsumerNames(ctx)
	var names []string
	for name := range lister.Name() {
		if strings.HasPrefix(name, "flow-") {
			names = append(names, name)
		}
	}
	if err := lister.Err(); err != nil {
		return nil, fmt.Errorf("list consumers: %w", err)
	}
	return names, nil
}

func (c *Client) ListConsumers(ctx context.Context, limit int, page int) ([]ConsumerSummary, uint64, error) {
	names, err := c.ListFlowConsumerNames(ctx)
	if err != nil {
		return nil, 0, err
	}
	total := uint64(len(names))
	names = paginate(names, limit, page)

	stream, err := c.js.Stream(ctx, c.streamName)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind to stream %s: %w", c.streamName, err)
	}

	result := make([]ConsumerSummary, 0, len(names))
	for _, name := range names {
		consumer, err := stream.Consumer(ctx, name)
		if err != nil {
			return nil, 0, fmt.Errorf("get consumer %s: %w", name, err)
		}
		info, err := consumer.Info(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("get consumer info %s: %w", name, err)
		}
		filters := info.Config.FilterSubjects
		if len(filters) == 0 && info.Config.FilterSubject != "" {
			filters = []string{info.Config.FilterSubject}
		}
		result = append(result, ConsumerSummary{
			Name:               name,
			FilterSubjects:     filters,
			NumPending:         info.NumPending,
			NumAckPending:      uint64(info.NumAckPending),
			DeliveredStreamSeq: info.Delivered.Stream,
			AckFloorStreamSeq:  info.AckFloor.Stream,
		})
	}
	return result, total, nil
}

// GetConsumerInfo returns the current sequence floor and pending message count.
// If name is empty, it aggregates across all flow-* consumers.
func (c *Client) GetConsumerInfo(ctx context.Context, name string) (uint64, uint64, error) {
	if strings.TrimSpace(name) != "" {
		consumer, err := c.GetConsumer(ctx, name)
		if err != nil {
			return 0, 0, err
		}

		info, err := consumer.Info(ctx)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to fetch consumer info for %s: %w", name, err)
		}
		return info.AckFloor.Stream, info.NumPending, nil
	}

	names, err := c.ListFlowConsumerNames(ctx)
	if err != nil {
		return 0, 0, err
	}
	if len(names) == 0 {
		return 0, 0, fmt.Errorf("no flow consumers found")
	}

	var ackFloor uint64
	var pending uint64
	matched := false
	for _, consumerName := range names {
		consumer, err := c.GetConsumer(ctx, consumerName)
		if err != nil {
			continue
		}
		info, err := consumer.Info(ctx)
		if err != nil {
			continue
		}
		if !matched || info.AckFloor.Stream < ackFloor {
			ackFloor = info.AckFloor.Stream
		}
		pending += info.NumPending
		matched = true
	}
	if !matched {
		return 0, 0, fmt.Errorf("no matching consumer info found")
	}
	return ackFloor, pending, nil
}
