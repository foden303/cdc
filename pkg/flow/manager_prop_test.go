package flow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/foden/cdc/pkg/interfaces"
	"github.com/foden/cdc/pkg/models"
	"github.com/nats-io/nats.go/jetstream"
	"pgregory.net/rapid"
)

// --- Mock implementations for testing ---

// mockStore is a simple in-memory store implementing interfaces.Store.
type mockStore struct {
	sources map[string]*interfaces.SourceConfig
	sinks   map[string]*interfaces.SinkConfig
	flows   map[string]*interfaces.FlowConfig
	offsets map[string]string
}

func newMockStore() *mockStore {
	return &mockStore{
		sources: make(map[string]*interfaces.SourceConfig),
		sinks:   make(map[string]*interfaces.SinkConfig),
		flows:   make(map[string]*interfaces.FlowConfig),
		offsets: make(map[string]string),
	}
}

func (s *mockStore) PutSource(_ context.Context, cfg *interfaces.SourceConfig) error {
	s.sources[cfg.InstanceID] = cfg
	return nil
}
func (s *mockStore) GetSource(_ context.Context, id string) (*interfaces.SourceConfig, error) {
	return s.sources[id], nil
}
func (s *mockStore) DeleteSource(_ context.Context, id string) error {
	delete(s.sources, id)
	return nil
}
func (s *mockStore) ListSources(_ context.Context) ([]*interfaces.SourceConfig, error) {
	var result []*interfaces.SourceConfig
	for _, v := range s.sources {
		result = append(result, v)
	}
	return result, nil
}
func (s *mockStore) PutSink(_ context.Context, cfg *interfaces.SinkConfig) error {
	s.sinks[cfg.InstanceID] = cfg
	return nil
}
func (s *mockStore) GetSink(_ context.Context, id string) (*interfaces.SinkConfig, error) {
	return s.sinks[id], nil
}
func (s *mockStore) DeleteSink(_ context.Context, id string) error {
	delete(s.sinks, id)
	return nil
}
func (s *mockStore) ListSinks(_ context.Context) ([]*interfaces.SinkConfig, error) {
	var result []*interfaces.SinkConfig
	for _, v := range s.sinks {
		result = append(result, v)
	}
	return result, nil
}
func (s *mockStore) PutFlow(_ context.Context, cfg *interfaces.FlowConfig) error {
	s.flows[cfg.FlowID] = cfg
	return nil
}
func (s *mockStore) GetFlow(_ context.Context, id string) (*interfaces.FlowConfig, error) {
	return s.flows[id], nil
}
func (s *mockStore) DeleteFlow(_ context.Context, id string) error {
	delete(s.flows, id)
	return nil
}
func (s *mockStore) ListFlows(_ context.Context) ([]*interfaces.FlowConfig, error) {
	var result []*interfaces.FlowConfig
	for _, v := range s.flows {
		result = append(result, v)
	}
	return result, nil
}
func (s *mockStore) SaveOffset(_ context.Context, flowID string, offset string) error {
	s.offsets[flowID] = offset
	return nil
}
func (s *mockStore) GetOffset(_ context.Context, flowID string) (string, error) {
	return s.offsets[flowID], nil
}

// mockRegistry implements interfaces.Registry for testing.
type mockRegistry struct {
	source interfaces.Source
}

func (r *mockRegistry) RegisterSource(_ string, _ interfaces.SourceFactory) {}
func (r *mockRegistry) RegisterSink(_ string, _ interfaces.SinkFactory)     {}
func (r *mockRegistry) CreateSource(_ *interfaces.SourceConfig) (interfaces.Source, error) {
	if r.source != nil {
		return r.source, nil
	}
	return &mockSource{}, nil
}
func (r *mockRegistry) CreateSink(_ *interfaces.SinkConfig) (interfaces.Sink, error) {
	return &mockSink{}, nil
}
func (r *mockRegistry) SourceNames() []string { return nil }
func (r *mockRegistry) SinkNames() []string   { return nil }

// mockSink implements interfaces.Sink for testing.
type mockSink struct{}

func (s *mockSink) WriteBatch(_ []*models.Event) error { return nil }
func (s *mockSink) Close() error                       { return nil }
func (s *mockSink) InstanceID() string                 { return "mock-sink" }
func (s *mockSink) Type() string                       { return "mock" }

// mockNATSClient implements interfaces.NATSClient for testing.
type mockNATSClient struct {
	mu        sync.Mutex
	published []*models.Event
	publishCh chan []*models.Event
}

func (n *mockNATSClient) PublishBatch(_ context.Context, _ func(*models.Event) string, events []*models.Event) error {
	n.mu.Lock()
	n.published = append(n.published, events...)
	n.mu.Unlock()
	if n.publishCh != nil {
		n.publishCh <- events
	}
	return nil
}
func (n *mockNATSClient) CreateOrUpdateConsumer(_ context.Context, _ string, _ []string) (jetstream.Consumer, error) {
	return &mockConsumer{}, nil
}
func (n *mockNATSClient) MoveToDLQ(_ context.Context, _ jetstream.Msg, _ interfaces.DLQMoveOptions) error {
	return nil
}
func (n *mockNATSClient) ReprocessDLQ(_ context.Context) (int, error) {
	return 0, nil
}
func (n *mockNATSClient) ListMessages(_ context.Context, _ models.MessageStatus, _ int, _ int, _ string, _ string) ([]*interfaces.NATSMessageItem, uint64, error) {
	return nil, 0, nil
}
func (n *mockNATSClient) ListDLQMessages(_ context.Context, _ int, _ int) ([]*interfaces.NATSMessageItem, uint64, error) {
	return nil, 0, nil
}
func (n *mockNATSClient) ListTopics(_ context.Context, _ int, _ int) ([]string, uint64, error) {
	return nil, 0, nil
}
func (n *mockNATSClient) ListPartitions(_ context.Context, _ string, _ int, _ int) ([]string, uint64, error) {
	return nil, 0, nil
}
func (n *mockNATSClient) ListConsumers(_ context.Context, _ int, _ int) ([]interfaces.NATSConsumerSummary, uint64, error) {
	return nil, 0, nil
}
func (n *mockNATSClient) CreateStream(_ context.Context, _ []string) error { return nil }
func (n *mockNATSClient) CreateDLQStream(_ context.Context) error          { return nil }
func (n *mockNATSClient) Close()                                           {}

type mockSource struct {
	started chan struct{}
	stopCh  chan struct{}
	events  []*models.Event
}

func (s *mockSource) Start(events chan<- *models.Event, _ <-chan uint64, _ string) error {
	if s.started != nil {
		close(s.started)
	}
	for _, ev := range s.events {
		events <- ev
	}
	return nil
}
func (s *mockSource) Stop() error {
	if s.stopCh != nil {
		close(s.stopCh)
	}
	return nil
}
func (s *mockSource) InstanceID() string               { return "src-1" }
func (s *mockSource) RegisterTable(_, _ string, _ int) {}
func (s *mockSource) UnregisterTable(_, _ string)      {}

// mockConsumer implements jetstream.Consumer for testing.
type mockConsumer struct{}

func (c *mockConsumer) Fetch(batch int, opts ...jetstream.FetchOpt) (jetstream.MessageBatch, error) {
	return &mockMessageBatch{}, nil
}
func (c *mockConsumer) FetchBytes(maxBytes int, opts ...jetstream.FetchOpt) (jetstream.MessageBatch, error) {
	return &mockMessageBatch{}, nil
}
func (c *mockConsumer) FetchNoWait(batch int) (jetstream.MessageBatch, error) {
	return &mockMessageBatch{}, nil
}
func (c *mockConsumer) Consume(handler jetstream.MessageHandler, opts ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error) {
	return nil, nil
}
func (c *mockConsumer) Messages(opts ...jetstream.PullMessagesOpt) (jetstream.MessagesContext, error) {
	return nil, nil
}
func (c *mockConsumer) Next(opts ...jetstream.FetchOpt) (jetstream.Msg, error) {
	return nil, context.DeadlineExceeded
}
func (c *mockConsumer) Info(ctx context.Context) (*jetstream.ConsumerInfo, error) {
	return &jetstream.ConsumerInfo{}, nil
}
func (c *mockConsumer) CachedInfo() *jetstream.ConsumerInfo {
	return &jetstream.ConsumerInfo{}
}

// mockMessageBatch implements jetstream.MessageBatch for testing.
type mockMessageBatch struct{}

func (b *mockMessageBatch) Messages() <-chan jetstream.Msg {
	ch := make(chan jetstream.Msg)
	close(ch)
	return ch
}
func (b *mockMessageBatch) Error() error { return nil }

// mockDiscovery implements interfaces.Discovery for testing.
type mockDiscovery struct{}

func (d *mockDiscovery) TestSourceConnection(_ context.Context, _ *interfaces.SourceConfig) (int64, error) {
	return 0, nil
}
func (d *mockDiscovery) TestSinkConnection(_ context.Context, _ *interfaces.SinkConfig) (int64, error) {
	return 0, nil
}
func (d *mockDiscovery) DiscoverSourceTables(_ context.Context, _ *interfaces.SourceConfig) ([]interfaces.TableInfo, error) {
	return nil, nil
}
func (d *mockDiscovery) DiscoverSinkTables(_ context.Context, _ *interfaces.SinkConfig) ([]interfaces.TableInfo, error) {
	return nil, nil
}

// --- Helper to create a test Manager ---

func newTestManager(store *mockStore) *Manager {
	pm := NewPoolManager()
	return NewManager(store, pm, &mockRegistry{}, &mockNATSClient{}, &mockDiscovery{})
}

func TestCreateFlowStartsSourceAndPublishesEvents(t *testing.T) {
	store := newMockStore()
	store.sources["src-1"] = &interfaces.SourceConfig{InstanceID: "src-1", Type: "postgres"}
	store.sinks["sink-1"] = &interfaces.SinkConfig{InstanceID: "sink-1", Type: "postgres"}

	started := make(chan struct{})
	src := &mockSource{
		started: started,
		events: []*models.Event{{
			Subject:    "cdc.src-1.public.users.0",
			InstanceID: "src-1",
			Schema:     "public",
			Table:      "users",
			Data:       []byte(`{"after":{"id":1}}`),
		}},
	}
	nc := &mockNATSClient{publishCh: make(chan []*models.Event, 1)}
	mgr := NewManager(store, NewPoolManager(), &mockRegistry{source: src}, nc, &mockDiscovery{})
	defer mgr.Stop()

	_, err := mgr.CreateFlow(context.Background(), &interfaces.FlowConfig{
		Name:        "users",
		SourceID:    "src-1",
		SinkID:      "sink-1",
		SourceTable: "public.users",
		SinkTable:   "users",
	})
	if err != nil {
		t.Fatalf("CreateFlow failed: %v", err)
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("source was not started")
	}

	select {
	case batch := <-nc.publishCh:
		if len(batch) != 1 {
			t.Fatalf("published batch length = %d, want 1", len(batch))
		}
		if batch[0].Subject != "cdc.src-1.public.users.0" {
			t.Fatalf("published subject = %q", batch[0].Subject)
		}
	case <-time.After(time.Second):
		t.Fatal("source event was not published to NATS")
	}
}

// createRunningFlow creates a flow in RUNNING state in the store.
func createRunningFlow(store *mockStore, flowID string) *interfaces.FlowConfig {
	flow := &interfaces.FlowConfig{
		FlowID:      flowID,
		Name:        "test-flow",
		SourceID:    "src-1",
		SinkID:      "sink-1",
		SourceTable: "public.users",
		SinkTable:   "users",
		Status:      interfaces.FlowStatusRunning,
	}
	store.flows[flowID] = flow
	return flow
}

// createPausedFlow creates a flow in PAUSED state in the store.
func createPausedFlow(store *mockStore, flowID string) *interfaces.FlowConfig {
	flow := &interfaces.FlowConfig{
		FlowID:      flowID,
		Name:        "test-flow",
		SourceID:    "src-1",
		SinkID:      "sink-1",
		SourceTable: "public.users",
		SinkTable:   "users",
		Status:      interfaces.FlowStatusPaused,
	}
	store.flows[flowID] = flow
	return flow
}

// --- Property Tests ---

// TestProperty_PauseRequiresRunning verifies that PauseFlow on a non-RUNNING flow
// returns ErrInvalidStateTransition.
// **Validates: Requirements 5.1, 5.4**
func TestProperty_PauseRequiresRunning(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		store := newMockStore()
		mgr := newTestManager(store)

		flowID := rapid.StringMatching(`[a-z0-9]{8}`).Draw(t, "flowID")
		// Pick a non-RUNNING status
		status := rapid.SampledFrom([]interfaces.FlowStatus{
			interfaces.FlowStatusPaused,
			interfaces.FlowStatusError,
		}).Draw(t, "status")

		store.flows[flowID] = &interfaces.FlowConfig{
			FlowID:      flowID,
			Name:        "test",
			SourceID:    "src-1",
			SinkID:      "sink-1",
			SourceTable: "public.t",
			SinkTable:   "t",
			Status:      status,
		}

		_, err := mgr.PauseFlow(context.Background(), flowID)
		if !errors.Is(err, ErrInvalidStateTransition) {
			t.Fatalf("expected ErrInvalidStateTransition for status=%q, got: %v", status, err)
		}
	})
}

// TestProperty_ResumeRequiresPaused verifies that ResumeFlow on a non-PAUSED flow
// returns ErrInvalidStateTransition.
// **Validates: Requirements 5.3, 5.5**
func TestProperty_ResumeRequiresPaused(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		store := newMockStore()
		mgr := newTestManager(store)

		// Need a sink in the store for resume to look up
		store.sinks["sink-1"] = &interfaces.SinkConfig{InstanceID: "sink-1", Type: "postgres"}

		flowID := rapid.StringMatching(`[a-z0-9]{8}`).Draw(t, "flowID")
		// Pick a non-PAUSED status
		status := rapid.SampledFrom([]interfaces.FlowStatus{
			interfaces.FlowStatusRunning,
			interfaces.FlowStatusError,
		}).Draw(t, "status")

		store.flows[flowID] = &interfaces.FlowConfig{
			FlowID:      flowID,
			Name:        "test",
			SourceID:    "src-1",
			SinkID:      "sink-1",
			SourceTable: "public.t",
			SinkTable:   "t",
			Status:      status,
		}

		_, err := mgr.ResumeFlow(context.Background(), flowID)
		if !errors.Is(err, ErrInvalidStateTransition) {
			t.Fatalf("expected ErrInvalidStateTransition for status=%q, got: %v", status, err)
		}
	})
}

// TestProperty_PauseResumeRoundTrip verifies that config is unchanged through
// a pause/resume cycle (name, source_id, sink_id, source_table, sink_table, column_mappings).
// **Validates: Requirements 5.1, 5.3**
func TestProperty_PauseResumeRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		store := newMockStore()
		mgr := newTestManager(store)

		// Set up required source and sink in store
		store.sources["src-1"] = &interfaces.SourceConfig{InstanceID: "src-1", Type: "postgres"}
		store.sinks["sink-1"] = &interfaces.SinkConfig{InstanceID: "sink-1", Type: "postgres"}

		flowID := rapid.StringMatching(`[a-z0-9]{8}`).Draw(t, "flowID")
		name := rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "name")
		sourceTable := rapid.StringMatching(`[a-z]{1,5}\.[a-z]{1,5}`).Draw(t, "sourceTable")
		sinkTable := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "sinkTable")

		// Generate some column mappings
		numMappings := rapid.IntRange(0, 5).Draw(t, "numMappings")
		var mappings []interfaces.ColumnMapping
		for i := 0; i < numMappings; i++ {
			mappings = append(mappings, interfaces.ColumnMapping{
				SourceColumn: rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "srcCol"),
				SinkColumn:   rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "sinkCol"),
				Enabled:      rapid.Bool().Draw(t, "enabled"),
			})
		}

		// Create a RUNNING flow directly in the store
		originalFlow := &interfaces.FlowConfig{
			FlowID:         flowID,
			Name:           name,
			SourceID:       "src-1",
			SinkID:         "sink-1",
			SourceTable:    sourceTable,
			SinkTable:      sinkTable,
			Status:         interfaces.FlowStatusRunning,
			ColumnMappings: mappings,
		}
		store.flows[flowID] = originalFlow

		// Pause the flow
		pausedFlow, err := mgr.PauseFlow(context.Background(), flowID)
		if err != nil {
			t.Fatalf("PauseFlow failed: %v", err)
		}
		if pausedFlow.Status != interfaces.FlowStatusPaused {
			t.Fatalf("expected PAUSED status, got %q", pausedFlow.Status)
		}

		// Resume the flow
		resumedFlow, err := mgr.ResumeFlow(context.Background(), flowID)
		if err != nil {
			t.Fatalf("ResumeFlow failed: %v", err)
		}
		if resumedFlow.Status != interfaces.FlowStatusRunning {
			t.Fatalf("expected RUNNING status after resume, got %q", resumedFlow.Status)
		}

		// Verify config unchanged through the cycle
		if resumedFlow.Name != name {
			t.Fatalf("name changed: %q -> %q", name, resumedFlow.Name)
		}
		if resumedFlow.SourceID != "src-1" {
			t.Fatalf("source_id changed: %q -> %q", "src-1", resumedFlow.SourceID)
		}
		if resumedFlow.SinkID != "sink-1" {
			t.Fatalf("sink_id changed: %q -> %q", "sink-1", resumedFlow.SinkID)
		}
		if resumedFlow.SourceTable != sourceTable {
			t.Fatalf("source_table changed: %q -> %q", sourceTable, resumedFlow.SourceTable)
		}
		if resumedFlow.SinkTable != sinkTable {
			t.Fatalf("sink_table changed: %q -> %q", sinkTable, resumedFlow.SinkTable)
		}
		if len(resumedFlow.ColumnMappings) != len(mappings) {
			t.Fatalf("column_mappings count changed: %d -> %d", len(mappings), len(resumedFlow.ColumnMappings))
		}
		for i, m := range mappings {
			rm := resumedFlow.ColumnMappings[i]
			if m.SourceColumn != rm.SourceColumn || m.SinkColumn != rm.SinkColumn || m.Enabled != rm.Enabled {
				t.Fatalf("column_mapping[%d] changed: %+v -> %+v", i, m, rm)
			}
		}
	})
}
