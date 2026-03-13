package sink_consumer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/foden/cdc/pkg/interfaces"
	"github.com/foden/cdc/pkg/queue"
	"github.com/foden/cdc/pkg/registry"
)

// Manager manages multiple sink consumers
type Manager struct {
	consumers   map[string]*SinkConsumer
	broker      *queue.Broker
	coordinator *queue.Coordinator
	mu          sync.RWMutex
}

// NewManager creates a new sink consumer manager
func NewManager(broker *queue.Broker, coordinator *queue.Coordinator) *Manager {
	return &Manager{
		consumers:   make(map[string]*SinkConsumer),
		broker:      broker,
		coordinator: coordinator,
	}
}

// Create creates and starts a new sink consumer
func (m *Manager) Create(ctx context.Context, config CreateConsumerConfig) (*SinkConsumer, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("consumer name is required")
	}
	if config.TopicName == "" {
		return nil, fmt.Errorf("topic_name is required")
	}
	if len(config.Sinks) == 0 {
		return nil, fmt.Errorf("at least one sink is required")
	}

	m.mu.Lock()

	// Generate ID if not provided
	id := fmt.Sprintf("sink-%d", len(m.consumers))
	if _, exists := m.consumers[id]; exists {
		m.mu.Unlock()
		return nil, fmt.Errorf("consumer with id %s already exists", id)
	}

	m.mu.Unlock()

	// Create sink instances
	sinks := make([]interfaces.Sink, len(config.Sinks))
	for i, spec := range config.Sinks {
		sink, err := registry.CreateSink(spec.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create sink %s: %w", spec.Type, err)
		}
		sinks[i] = sink
	}

	// Create consumer
	consumer, err := NewSinkConsumer(
		id,
		config.Name,
		config.TopicName,
		sinks,
		m.broker,
		m.coordinator,
		config,
	)
	if err != nil {
		for _, sink := range sinks {
			sink.Close()
		}
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Start consumer
	if err := consumer.Start(ctx); err != nil {
		for _, sink := range sinks {
			sink.Close()
		}
		return nil, fmt.Errorf("failed to start consumer: %w", err)
	}

	// Register consumer
	m.mu.Lock()
	m.consumers[id] = consumer
	m.mu.Unlock()

	slog.Info("sink consumer created and started",
		"consumer_id", id,
		"name", config.Name,
		"topic", config.TopicName)

	return consumer, nil
}

// Get retrieves a consumer by ID
func (m *Manager) Get(id string) *SinkConsumer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consumers[id]
}

// List returns all consumers
func (m *Manager) List() []*SinkConsumer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	consumers := make([]*SinkConsumer, 0, len(m.consumers))
	for _, c := range m.consumers {
		consumers = append(consumers, c)
	}
	return consumers
}

// Delete stops and removes a consumer
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	consumer, exists := m.consumers[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("consumer %s not found", id)
	}
	delete(m.consumers, id)
	m.mu.Unlock()

	// Stop consumer
	if err := consumer.Stop(); err != nil {
		slog.Error("failed to stop consumer", "consumer_id", id, "err", err)
		return err
	}

	slog.Info("sink consumer deleted", "consumer_id", id)
	return nil
}

// Pause pauses a consumer
func (m *Manager) Pause(id string) error {
	consumer := m.Get(id)
	if consumer == nil {
		return fmt.Errorf("consumer %s not found", id)
	}
	consumer.Pause()
	return nil
}

// Resume resumes a consumer
func (m *Manager) Resume(id string) error {
	consumer := m.Get(id)
	if consumer == nil {
		return fmt.Errorf("consumer %s not found", id)
	}
	consumer.Resume()
	return nil
}

// Stop stops all consumers gracefully
func (m *Manager) Stop() {
	m.mu.Lock()
	consumers := make([]*SinkConsumer, 0, len(m.consumers))
	for _, c := range m.consumers {
		consumers = append(consumers, c)
	}
	m.mu.Unlock()

	slog.Info("stopping all sink consumers", "count", len(consumers))

	// Stop all consumers in parallel
	var wg sync.WaitGroup
	for _, consumer := range consumers {
		wg.Add(1)
		go func(c *SinkConsumer) {
			defer wg.Done()
			if err := c.Stop(); err != nil {
				slog.Error("failed to stop consumer", "consumer_id", c.ID, "err", err)
			}
		}(consumer)
	}

	wg.Wait()
	slog.Info("all sink consumers stopped")
}
