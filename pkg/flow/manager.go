// Package flow provides the Flow Manager for orchestrating CDC flow lifecycle operations.
package flow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/foden/cdc/pkg/interfaces"
	"github.com/foden/cdc/pkg/models"
	"github.com/foden/cdc/pkg/registry"
	"github.com/google/uuid"
)

// FlowSink is the minimal sink interface needed by flow workers.
// This avoids a circular dependency with the interfaces package.
type FlowSink interface {
	WriteBatch(events []*models.Event) error
	Close() error
	InstanceID() string
}

// SinkProvider is a callback that resolves a sink instance by its ID.
// Used by the flow manager to obtain sink instances for flow workers.
type SinkProvider func(sinkID string) FlowSink

// FlowStatus represents the runtime state of a flow.
type FlowStatus = interfaces.FlowStatus

const (
	FlowStatusRunning = interfaces.FlowStatusRunning
	FlowStatusPaused  = interfaces.FlowStatusPaused
	FlowStatusError   = interfaces.FlowStatusError
)

// FlowOptions is a type alias for interfaces.FlowOptions.
type FlowOptions = interfaces.FlowOptions

// FlowConfig is a type alias for interfaces.FlowConfig.
type FlowConfig = interfaces.FlowConfig

// FlowStats is a type alias for interfaces.FlowStats.
type FlowStats = interfaces.FlowStats

// ErrFlowNotFound is returned when a flow cannot be found.
var ErrFlowNotFound = errors.New("flow not found")

// ErrInvalidStateTransition is returned when an invalid state transition is attempted.
var ErrInvalidStateTransition = errors.New("invalid state transition")

// Compile-time check that Manager implements interfaces.FlowManager.
var _ interfaces.FlowManager = (*Manager)(nil)

// Manager orchestrates flow lifecycle with Pool Manager integration.
// It implements interfaces.FlowManager.
type Manager struct {
	store       interfaces.Store
	poolManager *PoolManager
	sinkPool    *SinkPoolManager
	registry    interfaces.Registry
	natsClient  interfaces.NATSClient
	discovery   interfaces.Discovery
	maxDeliver  int
	mu          sync.RWMutex
	workers     map[string]*FlowWorker
	sources     map[string]interfaces.Source // source_id → running source instance
	sourceRuns  map[string]*sourceRuntime
	log         *slog.Logger
}

type sourceRuntime struct {
	source interfaces.Source
	events chan *models.Event
	acks   chan uint64
	cancel context.CancelFunc
	done   chan struct{}
}

// NewManager creates a new flow Manager with all dependencies.
func NewManager(
	store interfaces.Store,
	poolManager *PoolManager,
	registry interfaces.Registry,
	natsClient interfaces.NATSClient,
	discovery interfaces.Discovery,
	options ...ManagerOption,
) *Manager {
	m := &Manager{
		store:       store,
		poolManager: poolManager,
		sinkPool:    NewSinkPoolManager(registry, store),
		registry:    registry,
		natsClient:  natsClient,
		discovery:   discovery,
		maxDeliver:  defaultMaxDeliver,
		workers:     make(map[string]*FlowWorker),
		sources:     make(map[string]interfaces.Source),
		sourceRuns:  make(map[string]*sourceRuntime),
		log:         slog.With("component", "flow_manager"),
	}
	for _, option := range options {
		option(m)
	}
	if m.maxDeliver <= 0 {
		m.maxDeliver = defaultMaxDeliver
	}
	return m
}

type ManagerOption func(*Manager)

func WithMaxDeliver(maxDeliver int) ManagerOption {
	return func(m *Manager) {
		if maxDeliver > 0 {
			m.maxDeliver = maxDeliver
		}
	}
}

// CreateFlow validates refs, persists config, starts worker with ants pool.
func (m *Manager) CreateFlow(ctx context.Context, cfg *interfaces.FlowConfig) (*interfaces.FlowConfig, error) {
	// Validate required fields
	if cfg.Name == "" {
		return nil, fmt.Errorf("flow name is required")
	}
	if cfg.SourceID == "" {
		return nil, fmt.Errorf("source_id is required")
	}
	if cfg.SinkID == "" {
		return nil, fmt.Errorf("sink_id is required")
	}
	if cfg.SourceTable == "" {
		return nil, fmt.Errorf("source_table is required")
	}
	if cfg.SinkTable == "" {
		return nil, fmt.Errorf("sink_table is required")
	}

	// Validate source_id exists
	srcCfg, err := m.store.GetSource(ctx, cfg.SourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to look up source: %w", err)
	}
	if srcCfg == nil {
		return nil, fmt.Errorf("source %q not found", cfg.SourceID)
	}

	// Validate sink_id exists
	sinkCfg, err := m.store.GetSink(ctx, cfg.SinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to look up sink: %w", err)
	}
	if sinkCfg == nil {
		return nil, fmt.Errorf("sink %q not found", cfg.SinkID)
	}

	// Set default FlowOptions
	if cfg.Options == nil {
		cfg.Options = &interfaces.FlowOptions{}
	}
	if cfg.Options.PartitionCount <= 0 {
		cfg.Options.PartitionCount = 4
	}
	if cfg.Options.PoolSize <= 0 {
		cfg.Options.PoolSize = cfg.Options.PartitionCount
	}

	// Auto-generate column mappings if not provided
	if len(cfg.ColumnMappings) == 0 && m.discovery != nil {
		sourceTables, err := m.discovery.DiscoverSourceTables(ctx, srcCfg)
		if err == nil {
			sinkTables, err := m.discovery.DiscoverSinkTables(ctx, sinkCfg)
			if err == nil {
				sourceColumns := findTableColumns(sourceTables, cfg.SourceTable)
				sinkColumns := findTableColumns(sinkTables, cfg.SinkTable)
				if len(sourceColumns) > 0 && len(sinkColumns) > 0 {
					cfg.ColumnMappings = AutoGenerateMappings(sourceColumns, sinkColumns)
				}
			}
		}
	}

	// Generate flow_id (short UUID)
	flowID := uuid.New().String()[:8]
	now := time.Now().UnixMilli()

	cfg.FlowID = flowID
	cfg.Status = interfaces.FlowStatusRunning
	cfg.CreatedAt = now
	cfg.UpdatedAt = now

	// Persist to store
	if err := m.store.PutFlow(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to persist flow config: %w", err)
	}

	// Acquire shared sink instance from SinkPoolManager
	sink, err := m.sinkPool.Acquire(ctx, cfg.SinkID)
	if err != nil {
		m.log.Error("failed to acquire sink instance", "flow_id", flowID, "sink_id", cfg.SinkID, "err", err)
		return cfg, nil
	}

	m.startWorker(cfg, sink)
	if err := m.ensureSourceRunning(ctx, srcCfg); err != nil {
		m.log.Error("failed to start source",
			"flow_id", flowID,
			"source_id", cfg.SourceID,
			"err", err)
	}

	m.log.Info("flow created",
		"flow_id", flowID,
		"name", cfg.Name,
		"source_id", cfg.SourceID,
		"sink_id", cfg.SinkID,
		"source_table", cfg.SourceTable,
		"sink_table", cfg.SinkTable,
	)

	return cfg, nil
}

// GetFlow retrieves a single flow config from the store.
func (m *Manager) GetFlow(ctx context.Context, flowID string) (*interfaces.FlowConfig, error) {
	if flowID == "" {
		return nil, fmt.Errorf("flow_id is required")
	}

	flow, err := m.store.GetFlow(ctx, flowID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrFlowNotFound
	}

	return flow, nil
}

// ListFlows retrieves all flow configs from the store.
func (m *Manager) ListFlows(ctx context.Context) ([]*interfaces.FlowConfig, error) {
	return m.store.ListFlows(ctx)
}

// UpdateFlow applies changes and restarts the worker with new config.
func (m *Manager) UpdateFlow(ctx context.Context, cfg *interfaces.FlowConfig) (*interfaces.FlowConfig, error) {
	if cfg.FlowID == "" {
		return nil, fmt.Errorf("flow_id is required")
	}

	existing, err := m.store.GetFlow(ctx, cfg.FlowID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFlowNotFound
	}

	// Apply updates to existing config
	if cfg.Name != "" {
		existing.Name = cfg.Name
	}
	if cfg.SourceTable != "" {
		existing.SourceTable = cfg.SourceTable
	}
	if cfg.SinkTable != "" {
		existing.SinkTable = cfg.SinkTable
	}
	if cfg.ColumnMappings != nil {
		existing.ColumnMappings = cfg.ColumnMappings
	}
	if cfg.Options != nil {
		existing.Options = cfg.Options
	}
	existing.UpdatedAt = time.Now().UnixMilli()

	// Persist updated config
	if err := m.store.PutFlow(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to persist updated flow config: %w", err)
	}

	// Restart worker if flow is running
	if existing.Status == interfaces.FlowStatusRunning {
		m.stopWorker(existing.FlowID)
		m.poolManager.ReleasePool(existing.FlowID)
		m.sinkPool.Release(existing.SinkID)

		sink, err := m.sinkPool.Acquire(ctx, existing.SinkID)
		if err == nil {
			m.startWorker(existing, sink)
			if srcCfg, getErr := m.store.GetSource(ctx, existing.SourceID); getErr == nil && srcCfg != nil {
				if startErr := m.ensureSourceRunning(ctx, srcCfg); startErr != nil {
					m.log.Error("failed to start source on update", "flow_id", existing.FlowID, "err", startErr)
				}
			}
		} else {
			m.log.Error("failed to acquire sink on update", "flow_id", existing.FlowID, "err", err)
		}
	}

	m.log.Info("flow updated", "flow_id", existing.FlowID)
	return existing, nil
}

// PauseFlow releases pool, stops consumer, sets status=PAUSED (only from RUNNING).
func (m *Manager) PauseFlow(ctx context.Context, flowID string) (*interfaces.FlowConfig, error) {
	if flowID == "" {
		return nil, fmt.Errorf("flow_id is required")
	}

	flow, err := m.store.GetFlow(ctx, flowID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrFlowNotFound
	}

	// Validate state transition: only RUNNING flows can be paused
	if flow.Status != interfaces.FlowStatusRunning {
		return nil, ErrInvalidStateTransition
	}

	// Stop the flow worker
	m.stopWorker(flowID)

	// Release pool via PoolManager
	m.poolManager.ReleasePool(flowID)

	// Release shared sink connection
	m.sinkPool.Release(flow.SinkID)

	// Update status
	flow.Status = interfaces.FlowStatusPaused
	flow.UpdatedAt = time.Now().UnixMilli()

	if err := m.store.PutFlow(ctx, flow); err != nil {
		return nil, fmt.Errorf("failed to persist paused flow config: %w", err)
	}

	m.log.Info("flow paused", "flow_id", flowID)
	return flow, nil
}

// ResumeFlow creates new pool, resumes from last offset, sets status=RUNNING (only from PAUSED).
func (m *Manager) ResumeFlow(ctx context.Context, flowID string) (*interfaces.FlowConfig, error) {
	if flowID == "" {
		return nil, fmt.Errorf("flow_id is required")
	}

	flow, err := m.store.GetFlow(ctx, flowID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrFlowNotFound
	}

	// Validate state transition: only PAUSED flows can be resumed
	if flow.Status != interfaces.FlowStatusPaused {
		return nil, ErrInvalidStateTransition
	}

	// Acquire shared sink instance from SinkPoolManager
	sink, err := m.sinkPool.Acquire(ctx, flow.SinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire sink instance: %w", err)
	}

	// Start new FlowWorker (resumes from last offset via store.GetOffset)
	m.startWorker(flow, sink)
	if srcCfg, err := m.store.GetSource(ctx, flow.SourceID); err == nil && srcCfg != nil {
		if err := m.ensureSourceRunning(ctx, srcCfg); err != nil {
			m.log.Error("failed to start source on resume", "flow_id", flow.FlowID, "err", err)
		}
	}

	// Update status
	flow.Status = interfaces.FlowStatusRunning
	flow.UpdatedAt = time.Now().UnixMilli()

	if err := m.store.PutFlow(ctx, flow); err != nil {
		return nil, fmt.Errorf("failed to persist resumed flow config: %w", err)
	}

	m.log.Info("flow resumed", "flow_id", flowID)
	return flow, nil
}

// DeleteFlow releases pool, deletes consumer, removes from KV.
func (m *Manager) DeleteFlow(ctx context.Context, flowID string) error {
	if flowID == "" {
		return fmt.Errorf("flow_id is required")
	}

	flow, err := m.store.GetFlow(ctx, flowID)
	if err != nil {
		return err
	}
	if flow == nil {
		return ErrFlowNotFound
	}

	// Stop worker if running
	m.stopWorker(flowID)

	// Release pool
	m.poolManager.ReleasePool(flowID)

	// Release shared sink connection
	m.sinkPool.Release(flow.SinkID)

	// Delete from store
	if err := m.store.DeleteFlow(ctx, flowID); err != nil {
		return fmt.Errorf("failed to delete flow from store: %w", err)
	}

	// Delete offset from store (save empty to clear)
	_ = m.store.SaveOffset(ctx, flowID, "")

	m.log.Info("flow deleted", "flow_id", flowID)
	return nil
}

// GetFlowStats returns per-flow metrics from Pool Manager.
func (m *Manager) GetFlowStats(ctx context.Context, flowID string) (*interfaces.FlowStats, error) {
	if flowID == "" {
		return nil, fmt.Errorf("flow_id is required")
	}

	flow, err := m.store.GetFlow(ctx, flowID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrFlowNotFound
	}

	stats := &interfaces.FlowStats{
		LastSyncedAt: flow.UpdatedAt,
	}

	// If flow is paused, return zeroed stats (no active pool)
	if flow.Status == interfaces.FlowStatusPaused {
		return stats, nil
	}

	// Get pool metrics for running flows
	metrics := m.poolManager.GetMetrics(flowID)
	if metrics != nil {
		// Pool metrics provide capacity and utilization info.
		// TotalEventsProcessed and EventsPerSecond would ideally come from
		// a dedicated metrics collector; for now we expose pool state.
		stats.TotalEventsProcessed = 0
		stats.ReplicationLagMs = 0
		stats.EventsPerSecond = 0
	}

	return stats, nil
}

// RestoreFlows loads all flows from KV on startup, starts RUNNING flows, skips PAUSED.
func (m *Manager) RestoreFlows(ctx context.Context) error {
	flows, err := m.store.ListFlows(ctx)
	if err != nil {
		return fmt.Errorf("failed to restore flows: %w", err)
	}

	restored := 0
	started := 0

	for _, flow := range flows {
		switch flow.Status {
		case interfaces.FlowStatusRunning:
			// Acquire shared sink and start worker
			sink, err := m.sinkPool.Acquire(ctx, flow.SinkID)
			if err != nil {
				m.log.Error("failed to acquire sink for flow restore",
					"flow_id", flow.FlowID, "sink_id", flow.SinkID, "err", err)
				continue
			}

			m.startWorker(flow, sink)
			if srcCfg, getErr := m.store.GetSource(ctx, flow.SourceID); getErr == nil && srcCfg != nil {
				if startErr := m.ensureSourceRunning(ctx, srcCfg); startErr != nil {
					m.log.Error("failed to start source on restore", "flow_id", flow.FlowID, "err", startErr)
				}
			}
			started++
			m.log.Info("flow restored and worker started", "flow_id", flow.FlowID, "name", flow.Name)

		case interfaces.FlowStatusPaused:
			// Just load config, no worker
			m.log.Info("flow restored without starting worker", "flow_id", flow.FlowID, "name", flow.Name, "status", flow.Status)
		}
		restored++
	}

	m.log.Info("flows restored from store", "total", restored, "started", started)
	return nil
}

// Stop gracefully stops all workers.
func (m *Manager) Stop() {
	m.mu.Lock()
	for flowID, worker := range m.workers {
		m.log.Info("stopping flow worker on shutdown", "flow_id", flowID)
		worker.Stop()
	}
	m.workers = make(map[string]*FlowWorker)

	for sourceID, runtime := range m.sourceRuns {
		m.log.Info("stopping source on shutdown", "source_id", sourceID)
		runtime.cancel()
		_ = runtime.source.Stop()
		<-runtime.done
	}
	m.sources = make(map[string]interfaces.Source)
	m.sourceRuns = make(map[string]*sourceRuntime)
	m.mu.Unlock()

	// Close all shared sink connections
	m.sinkPool.CloseAll()

	m.log.Info("flow manager stopped")
}

// --- Internal helpers ---

// RegisterSource registers a running source instance with the flow manager.
// This allows the flow manager to call RegisterTable/UnregisterTable on sources
// when flows start/stop.
func (m *Manager) RegisterSource(sourceID string, src interfaces.Source) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sources[sourceID] = src
}

// UnregisterSource removes a source instance from the flow manager.
func (m *Manager) UnregisterSource(sourceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sources, sourceID)
}

func (m *Manager) ensureSourceRunning(ctx context.Context, cfg *interfaces.SourceConfig) error {
	if cfg == nil {
		return fmt.Errorf("source config is required")
	}

	m.mu.RLock()
	if _, ok := m.sourceRuns[cfg.InstanceID]; ok {
		m.mu.RUnlock()
		return nil
	}
	m.mu.RUnlock()

	src, err := m.registry.CreateSource(cfg)
	if err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(context.Background())
	runtime := &sourceRuntime{
		source: src,
		events: make(chan *models.Event, 8192),
		acks:   make(chan uint64, 1024),
		cancel: cancel,
		done:   make(chan struct{}),
	}

	m.mu.Lock()
	if _, ok := m.sourceRuns[cfg.InstanceID]; ok {
		m.mu.Unlock()
		cancel()
		return nil
	}
	m.sourceRuns[cfg.InstanceID] = runtime
	m.sources[cfg.InstanceID] = src
	m.mu.Unlock()

	go m.publishSourceEvents(runCtx, cfg.InstanceID, runtime.events, runtime.done)

	if err := src.Start(runtime.events, runtime.acks, ""); err != nil {
		cancel()
		<-runtime.done
		m.mu.Lock()
		delete(m.sourceRuns, cfg.InstanceID)
		delete(m.sources, cfg.InstanceID)
		m.mu.Unlock()
		return err
	}

	m.log.Info("source started", "source_id", cfg.InstanceID, "type", cfg.Type)
	return nil
}

func (m *Manager) publishSourceEvents(ctx context.Context, sourceID string, events <-chan *models.Event, done chan<- struct{}) {
	defer close(done)

	batch := make([]*models.Event, 0, 100)
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		toPublish := append([]*models.Event(nil), batch...)
		batch = batch[:0]
		if err := m.natsClient.PublishBatch(ctx, func(ev *models.Event) string {
			return ev.Subject
		}, toPublish); err != nil {
			m.log.Error("failed to publish source events", "source_id", sourceID, "count", len(toPublish), "err", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case ev, ok := <-events:
			if !ok {
				flush()
				return
			}
			if ev == nil {
				continue
			}
			batch = append(batch, ev)
			if len(batch) >= 100 {
				flush()
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(100 * time.Millisecond)
			}
		case <-timer.C:
			flush()
			timer.Reset(100 * time.Millisecond)
		}
	}
}

// startWorker starts a flow worker for the given flow config with the provided sink.
func (m *Manager) startWorker(flow *interfaces.FlowConfig, sink interfaces.Sink) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing worker if any
	if existing, ok := m.workers[flow.FlowID]; ok {
		existing.Stop()
		delete(m.workers, flow.FlowID)
	}

	// Register table in global table registry
	schema, table := parseSourceTable(flow.SourceTable)
	partitionCount := 4
	if flow.Options != nil && flow.Options.PartitionCount > 0 {
		partitionCount = flow.Options.PartitionCount
	}
	registry.GlobalTableRegistry.Register(flow.SourceID, schema, table, partitionCount)

	// Wrap interfaces.Sink as FlowSink
	fs := &sinkAdapter{sink: sink}

	worker := StartFlowWorker(context.Background(), flow, fs, m.poolManager, m.store, m.natsClient, m.maxDeliver)
	m.workers[flow.FlowID] = worker
	m.log.Info("flow worker started", "flow_id", flow.FlowID, "sink_id", flow.SinkID)
}

// stopWorker stops the flow worker for the given flow ID.
func (m *Manager) stopWorker(flowID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if worker, ok := m.workers[flowID]; ok {
		// Unregister table from global table registry
		flow := worker.flow
		schema, table := parseSourceTable(flow.SourceTable)
		registry.GlobalTableRegistry.Unregister(flow.SourceID, schema, table)

		worker.Stop()
		delete(m.workers, flowID)
		m.log.Info("flow worker stopped", "flow_id", flowID)
	}
}

// findTableColumns finds columns for a specific table from a list of discovered tables.
func findTableColumns(tables []interfaces.TableInfo, tableName string) []interfaces.ColumnInfo {
	// tableName can be "schema.table" or just "table"
	for _, t := range tables {
		fullName := t.Name
		if t.Schema != "" {
			fullName = t.Schema + "." + t.Name
		}
		if strings.EqualFold(fullName, tableName) || strings.EqualFold(t.Name, tableName) {
			return t.Columns
		}
	}
	return nil
}

// parseSourceTable splits a "schema.table" string into schema and table parts.
func parseSourceTable(sourceTable string) (schema, table string) {
	parts := strings.SplitN(sourceTable, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", sourceTable
}

// sinkAdapter wraps interfaces.Sink to satisfy the FlowSink interface.
type sinkAdapter struct {
	sink interfaces.Sink
}

func (a *sinkAdapter) WriteBatch(events []*models.Event) error {
	return a.sink.WriteBatch(events)
}

func (a *sinkAdapter) Close() error {
	return a.sink.Close()
}

func (a *sinkAdapter) InstanceID() string {
	return a.sink.InstanceID()
}
