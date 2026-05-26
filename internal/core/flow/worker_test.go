package flow

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/internal/core/ports"
	coreruntime "github.com/foden/cdc/internal/core/runtime"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/panjf2000/ants/v2"
)

type workerTestMsg struct {
	meta  *jetstream.MsgMetadata
	naked bool
}

func (m *workerTestMsg) Metadata() (*jetstream.MsgMetadata, error) { return m.meta, nil }
func (m *workerTestMsg) Data() []byte                              { return []byte(`{"ok":true}`) }
func (m *workerTestMsg) Headers() nats.Header                      { return nats.Header{} }
func (m *workerTestMsg) Subject() string                           { return "cdc.src.public.users.0" }
func (m *workerTestMsg) Reply() string                             { return "" }
func (m *workerTestMsg) Ack() error                                { return nil }
func (m *workerTestMsg) DoubleAck(context.Context) error           { return nil }
func (m *workerTestMsg) Nak() error                                { m.naked = true; return nil }
func (m *workerTestMsg) NakWithDelay(time.Duration) error          { m.naked = true; return nil }
func (m *workerTestMsg) InProgress() error                         { return nil }
func (m *workerTestMsg) Term() error                               { return nil }
func (m *workerTestMsg) TermWithReason(string) error               { return nil }

type workerTestNATS struct {
	dlqMoves int
}

type failingPoolManager struct{}

func (f *failingPoolManager) CreatePool(string, int) (*ants.Pool, error) {
	return nil, errors.New("shared pool rejected")
}

func (f *failingPoolManager) ReleasePool(string) {}

func (n *workerTestNATS) PublishBatch(context.Context, func(*domain.Event) string, []*domain.Event) error {
	return nil
}
func (n *workerTestNATS) CreateOrUpdateConsumer(context.Context, string, []string) (jetstream.Consumer, error) {
	return nil, nil
}
func (n *workerTestNATS) MoveToDLQ(context.Context, jetstream.Msg, ports.DLQMoveOptions) error {
	n.dlqMoves++
	return nil
}
func (n *workerTestNATS) ReprocessDLQ(context.Context) (int, error) { return 0, nil }
func (n *workerTestNATS) ListMessages(context.Context, domain.MessageStatus, int, int, string, string) ([]*ports.NATSMessageItem, uint64, error) {
	return nil, 0, nil
}
func (n *workerTestNATS) ListDLQMessages(context.Context, int, int) ([]*ports.NATSMessageItem, uint64, error) {
	return nil, 0, nil
}
func (n *workerTestNATS) ListTopics(context.Context, int, int) ([]string, uint64, error) {
	return nil, 0, nil
}
func (n *workerTestNATS) ListPartitions(context.Context, string, int, int) ([]string, uint64, error) {
	return nil, 0, nil
}
func (n *workerTestNATS) ListConsumers(context.Context, int, int) ([]ports.NATSConsumerSummary, uint64, error) {
	return nil, 0, nil
}
func (n *workerTestNATS) CreateStream(context.Context, []string) error { return nil }
func (n *workerTestNATS) CreateDLQStream(context.Context) error        { return nil }
func (n *workerTestNATS) Health(context.Context) error                 { return nil }
func (n *workerTestNATS) Close()                                       {}

func TestStartFlowWorkerReturnsErrorWhenFallbackPoolCreationFails(t *testing.T) {
	originalNewAntsPool := newAntsPool
	newAntsPool = func(int, ...ants.Option) (*ants.Pool, error) {
		return nil, errors.New("isolated pool failed")
	}
	t.Cleanup(func() { newAntsPool = originalNewAntsPool })

	worker, err := StartFlowWorker(
		context.Background(),
		&FlowConfig{FlowID: "flow-1", SourceTable: "public.users", SinkTable: "public.users"},
		nil,
		&failingPoolManager{},
		nil,
		&workerTestNATS{},
		3,
		&coreruntime.Metrics{},
	)
	if err == nil {
		t.Fatal("expected pool creation error")
	}
	if worker != nil {
		t.Fatal("worker should be nil when pool creation fails")
	}
}

func TestHandleFailureUsesConfiguredMaxDeliver(t *testing.T) {
	natsClient := &workerTestNATS{}
	worker := &FlowWorker{
		flow:       &FlowConfig{FlowID: "flow-1", SinkID: "sink-1"},
		natsClient: natsClient,
		maxDeliver: 3,
		log:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	retryMsg := &workerTestMsg{meta: &jetstream.MsgMetadata{NumDelivered: 2}}
	worker.handleFailure(context.Background(), []jetstream.Msg{retryMsg})
	if !retryMsg.naked {
		t.Fatal("message below maxDeliver should be NAKed")
	}
	if natsClient.dlqMoves != 0 {
		t.Fatalf("DLQ moves = %d, want 0", natsClient.dlqMoves)
	}

	dlqMsg := &workerTestMsg{meta: &jetstream.MsgMetadata{NumDelivered: 3}}
	worker.handleFailure(context.Background(), []jetstream.Msg{dlqMsg})
	if natsClient.dlqMoves != 1 {
		t.Fatalf("DLQ moves = %d, want 1", natsClient.dlqMoves)
	}
}

func TestApplySinkTableUsesFlowSinkTable(t *testing.T) {
	event := &domain.Event{Schema: "public", Table: "users"}

	applySinkTable(event, "warehouse.customers")

	if event.Schema != "warehouse" || event.Table != "customers" {
		t.Fatalf("event table = %s.%s", event.Schema, event.Table)
	}
}

func TestApplySinkTableClearsSchemaWhenSinkTableUnqualified(t *testing.T) {
	event := &domain.Event{Schema: "public", Table: "users"}

	applySinkTable(event, "customers")

	if event.Schema != "" || event.Table != "customers" {
		t.Fatalf("event table = %s.%s", event.Schema, event.Table)
	}
}
