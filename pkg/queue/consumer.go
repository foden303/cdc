package queue

import (
	"context"
	"log/slog"
	"time"
)

type Consumer struct {
	ID string

	Group      *ConsumerGroup
	Broker     *Broker
	partitions []*Partition
	dlq        *DLQ
	retryTopic *Topic
}

func (c *Consumer) handleError(msg *MessageView) {
	if msg.Retry > 3 {
		if c.dlq != nil {
			c.dlq.Send(msg)
		} else {
			slog.Error("message exceeded max retries but DLQ not configured",
				"key", string(msg.Key), "offset", msg.Offset)
		}
		return
	}

	retryMsg := &Message{
		Key:       msg.Key,
		Value:     msg.Value,
		Timestamp: msg.Timestamp,
		Retry:     msg.Retry + 1,
	}

	if c.retryTopic != nil {
		c.retryTopic.Partition(msg.Key).Produce(retryMsg)
	} else {
		slog.Error("retry topic not configured, dropping message",
			"key", string(msg.Key), "offset", msg.Offset)
	}
}

func (c *Consumer) Poll(ctx context.Context, process func(*MessageView) error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		idle := true
		for _, part := range c.partitions {
			c.Group.mu.RLock()
			topicOffsets := c.Group.offsets[part.topic]
			var offset uint64
			if topicOffsets != nil {
				offset = topicOffsets[part.id]
			}
			c.Group.mu.RUnlock()

			msgs, err := part.Fetch(offset, 1<<20)
			if err != nil {
				continue
			}

			for _, m := range msgs {
				idle = false
				err := process(m)
				if err != nil {
					c.handleError(m)
					continue
				}

				c.Group.mu.Lock()
				if c.Group.offsets[part.topic] == nil {
					c.Group.offsets[part.topic] = make(map[int]uint64)
				}
				c.Group.offsets[part.topic][part.id] = m.Offset + 1
				c.Group.mu.Unlock()
			}
		}

		if idle {
			time.Sleep(100 * time.Millisecond)
		}
	}
}
