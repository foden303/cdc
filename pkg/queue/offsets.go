package queue

import (
	"fmt"
	"time"
)

func (c *Coordinator) CommitOffset(groupID string, topic string, partition int, offset uint64) error {
	c.mu.Lock()
	g, ok := c.groups[groupID]
	c.mu.Unlock()

	if !ok {
		return fmt.Errorf("group not found")
	}

	g.mu.Lock()
	if g.offsets[topic] == nil {
		g.offsets[topic] = make(map[int]uint64)
	}
	g.offsets[topic][partition] = offset
	g.mu.Unlock()

	// Persist to internal topic __consumer_offsets
	msg := &Message{
		Key:       []byte(fmt.Sprintf("%s-%d", groupID, partition)),
		Value:     []byte(fmt.Sprintf("%d", offset)),
		Timestamp: time.Now().UnixMilli(),
	}

	t := c.broker.Topic("__consumer_offsets")
	if t == nil {
		if err := c.broker.CreateTopic("__consumer_offsets", 50); err != nil {
			return fmt.Errorf("failed to create __consumer_offsets topic: %w", err)
		}
		t = c.broker.Topic("__consumer_offsets")
	}

	_, err := t.Partition(msg.Key).Produce(msg)
	return err
}

func (c *Coordinator) FetchOffset(groupID string, topic string, partition int) (uint64, error) {
	c.mu.Lock()
	g := c.groups[groupID]
	c.mu.Unlock()

	if g == nil {
		return 0, fmt.Errorf("group not found")
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.offsets[topic] == nil {
		return 0, nil
	}
	return g.offsets[topic][partition], nil
}
