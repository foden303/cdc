package flow

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/internal/core/ports"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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
