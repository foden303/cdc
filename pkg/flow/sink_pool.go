package flow

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/foden/cdc/pkg/interfaces"
)

// SinkPoolManager manages shared sink instances with reference counting.
// Multiple flows using the same sink_id share one connection pool.
type SinkPoolManager struct {
	mu       sync.Mutex
	registry interfaces.Registry
	store    interfaces.Store
	pools    map[string]*sinkEntry // sink_id → entry
	log      *slog.Logger
}

// sinkEntry holds a shared sink instance and its reference count.
type sinkEntry struct {
	sink     interfaces.Sink
	refCount int
}

// NewSinkPoolManager creates a new SinkPoolManager.
func NewSinkPoolManager(registry interfaces.Registry, store interfaces.Store) *SinkPoolManager {
	return &SinkPoolManager{
		registry: registry,
		store:    store,
		pools:    make(map[string]*sinkEntry),
		log:      slog.With("component", "sink_pool_manager"),
	}
}

// Acquire returns a shared sink instance for the given sink_id.
// Creates a new one if not exists, increments refCount.
func (m *SinkPoolManager) Acquire(ctx context.Context, sinkID string) (interfaces.Sink, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If already exists, increment refCount and return
	if entry, ok := m.pools[sinkID]; ok {
		entry.refCount++
		m.log.Debug("sink acquired (existing)", "sink_id", sinkID, "ref_count", entry.refCount)
		return entry.sink, nil
	}

	// Look up sink config from store
	sinkCfg, err := m.store.GetSink(ctx, sinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to look up sink %q: %w", sinkID, err)
	}
	if sinkCfg == nil {
		return nil, fmt.Errorf("sink %q not found", sinkID)
	}

	// Create new sink instance via registry
	sink, err := m.registry.CreateSink(sinkCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create sink %q: %w", sinkID, err)
	}

	m.pools[sinkID] = &sinkEntry{
		sink:     sink,
		refCount: 1,
	}

	m.log.Info("sink acquired (new)", "sink_id", sinkID)
	return sink, nil
}

// Release decrements refCount for a sink. Closes the sink if refCount reaches 0.
func (m *SinkPoolManager) Release(sinkID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.pools[sinkID]
	if !ok {
		return
	}

	entry.refCount--
	m.log.Debug("sink released", "sink_id", sinkID, "ref_count", entry.refCount)

	if entry.refCount <= 0 {
		if err := entry.sink.Close(); err != nil {
			m.log.Error("failed to close sink", "sink_id", sinkID, "err", err)
		}
		delete(m.pools, sinkID)
		m.log.Info("sink closed and removed from pool", "sink_id", sinkID)
	}
}

// CloseAll closes all managed sinks (for shutdown).
func (m *SinkPoolManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for sinkID, entry := range m.pools {
		if err := entry.sink.Close(); err != nil {
			m.log.Error("failed to close sink during shutdown", "sink_id", sinkID, "err", err)
		}
	}
	m.pools = make(map[string]*sinkEntry)
	m.log.Info("all sinks closed")
}
