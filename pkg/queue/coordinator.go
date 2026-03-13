package queue

import "sync"

type Coordinator struct {
	mu     sync.Mutex
	groups map[string]*ConsumerGroup
	broker *Broker
}

func NewCoordinator(broker *Broker) *Coordinator {
	return &Coordinator{
		groups: make(map[string]*ConsumerGroup),
		broker: broker,
	}
}

func (c *Coordinator) Join(groupID string, consumer *Consumer, topic string, partitions int) {
	c.mu.Lock()
	g, ok := c.groups[groupID]
	if !ok {
		g = &ConsumerGroup{
			ID:          groupID,
			consumers:   make(map[string]*Consumer),
			assignments: make(map[string]map[int]string),
			offsets:     make(map[string]map[int]uint64),
		}
		c.groups[groupID] = g
	}
	c.mu.Unlock()

	g.mu.Lock()
	g.consumers[consumer.ID] = consumer
	g.mu.Unlock()

	g.rebalance(topic, partitions)
}

func (g *ConsumerGroup) rebalance(topic string, partitions int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	consumers := make([]string, 0, len(g.consumers))
	for id := range g.consumers {
		consumers = append(consumers, id)
	}
	if len(consumers) == 0 {
		return
	}

	if g.assignments[topic] == nil {
		g.assignments[topic] = make(map[int]string)
	}

	for i := 0; i < partitions; i++ {
		consumerID := consumers[i%len(consumers)]
		g.assignments[topic][i] = consumerID
	}
}
