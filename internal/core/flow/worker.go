// Package flow provides the Flow Manager and FlowWorker for orchestrating CDC flow lifecycle operations.
package flow

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bytedance/sonic"

	"github.com/foden/cdc/internal/adapters/driven/metrics"
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/internal/core/ports"
	coreruntime "github.com/foden/cdc/internal/core/runtime"
	cdcerrors "github.com/foden/cdc/pkg/errors"
	"github.com/foden/cdc/pkg/pool"
	"github.com/foden/cdc/pkg/retry"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/panjf2000/ants/v2"
)

// defaultPoolSize is the default number of goroutines in a flow's ants pool.
const defaultPoolSize = 4

// defaultMaxDeliver matches the config package default for NATS max_deliver.
const defaultMaxDeliver = 5

var newAntsPool = ants.NewPool

// FlowWorker processes events for a single flow using a dedicated ants pool.
// Each FlowWorker owns a NATS durable consumer filtered to its source table
// and submits batch processing tasks to its ants pool.
type flowPoolManager interface {
	CreatePool(flowID string, size int) (*ants.Pool, error)
	ReleasePool(flowID string)
}

type FlowWorker struct {
	flow           *FlowConfig
	sink           FlowSink
	pool           *ants.Pool
	poolManager    flowPoolManager
	store          ports.Store
	natsClient     ports.NATSClient
	runtimeMetrics *coreruntime.Metrics
	filter         *Filter
	mappings       []ports.ColumnMapping
	maxDeliver     int
	log            *slog.Logger
	cancel         context.CancelFunc
	stopped        chan struct{}
}

// StartFlowWorker creates a NATS durable consumer and ants pool for the flow,
// then starts the main processing loop in a goroutine.
func StartFlowWorker(
	ctx context.Context,
	flow *FlowConfig,
	sink FlowSink,
	poolManager flowPoolManager,
	store ports.Store,
	natsClient ports.NATSClient,
	maxDeliver int,
	runtimeMetrics *coreruntime.Metrics,
) (*FlowWorker, error) {
	ctx, cancel := context.WithCancel(ctx)

	// Create component-scoped logger with flow context
	log := slog.With(
		"component", "flow_worker",
		"flow_id", flow.FlowID,
		"source_table", flow.SourceTable,
		"sink_table", flow.SinkTable,
	)

	// Parse and compile filter expression
	var filter *Filter
	if flow.Options != nil && flow.Options.FilterExpression != "" {
		var err error
		filter, err = NewFilter(flow.Options.FilterExpression)
		if err != nil {
			log.Error("failed to compile filter expression, using pass-all",
				"expression", flow.Options.FilterExpression,
				"err", err)
			filter = nil
		}
	}

	// Determine pool size: default to partition count for 1:1 ordering guarantee
	poolSize := defaultPoolSize
	if flow.Options != nil && flow.Options.PoolSize > 0 {
		poolSize = flow.Options.PoolSize
	} else if flow.Options != nil && flow.Options.PartitionCount > 0 {
		poolSize = flow.Options.PartitionCount
	}
	if maxDeliver <= 0 {
		maxDeliver = defaultMaxDeliver
	}

	// Get or create ants pool from PoolManager
	antsPool, err := poolManager.CreatePool(flow.FlowID, poolSize)
	if err != nil {
		log.Error("failed to create shared ants pool, using isolated pool",
			"pool_size", poolSize,
			"err", err)
		// Keep the worker isolated if the shared pool manager rejects this flow.
		antsPool, err = newAntsPool(poolSize)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("create isolated ants pool: %w", err)
		}
	}

	w := &FlowWorker{
		flow:           flow,
		sink:           sink,
		pool:           antsPool,
		poolManager:    poolManager,
		store:          store,
		natsClient:     natsClient,
		runtimeMetrics: runtimeMetrics,
		filter:         filter,
		mappings:       flow.ColumnMappings,
		maxDeliver:     maxDeliver,
		log:            log,
		cancel:         cancel,
		stopped:        make(chan struct{}),
	}

	go w.run(ctx)
	return w, nil
}

// flowConsumerName returns the NATS durable consumer name for this flow.
func flowConsumerName(flowID string) string {
	return fmt.Sprintf("flow-%s", flowID)
}

// run is the main processing loop: fetch batch from NATS → submit to ants pool.
func (w *FlowWorker) run(ctx context.Context) {
	defer close(w.stopped)

	// Parse source_table to get schema.table
	schema, table := parseSourceTable(w.flow.SourceTable)

	// Build filter subject: cdc.<source_id>.<schema>.<table>.*
	// This matches all partitions for the specific source table.
	filterSubject := fmt.Sprintf("cdc.%s.%s.%s.*", w.flow.SourceID, schema, table)
	consumerName := flowConsumerName(w.flow.FlowID)

	consumer, err := w.natsClient.CreateOrUpdateConsumer(ctx, consumerName, []string{filterSubject})
	if err != nil {
		w.log.Error("failed to create consumer",
			"consumer", consumerName,
			"filter", filterSubject,
			"err", err)
		return
	}

	// Determine batch size
	batchSize := 100
	if w.flow.Options != nil && w.flow.Options.BatchSize > 0 {
		batchSize = int(w.flow.Options.BatchSize)
	}

	// Determine flush interval (used as FetchMaxWait)
	flushInterval := time.Second
	if w.flow.Options != nil && w.flow.Options.FlushIntervalMs > 0 {
		flushInterval = time.Duration(w.flow.Options.FlushIntervalMs) * time.Millisecond
	}

	w.log.Info("worker started",
		"consumer", consumerName,
		"filter", filterSubject,
		"batch_size", batchSize,
		"flush_interval", flushInterval,
		"pool_size", w.pool.Cap())

	for {
		select {
		case <-ctx.Done():
			w.log.Info("worker stopping due to context cancellation")
			return
		default:
		}

		// Fetch a batch of messages from the NATS consumer
		msgBatch, err := consumer.Fetch(batchSize, jetstream.FetchMaxWait(flushInterval))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			// Transient fetch error — back off briefly
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Collect messages from the batch channel
		msgs := make([]jetstream.Msg, 0, batchSize)
		for msg := range msgBatch.Messages() {
			msgs = append(msgs, msg)
		}

		if len(msgs) == 0 {
			continue
		}

		// Submit batch processing as a task to the ants pool
		batchMsgs := msgs // capture for closure
		err = w.pool.Submit(func() {
			w.processBatch(ctx, batchMsgs)
		})
		if err != nil {
			w.log.Error("failed to submit task to pool, NAKing batch",
				"batch_size", len(batchMsgs),
				"err", err)
			// NAK all messages so they can be redelivered
			for _, msg := range batchMsgs {
				_ = msg.Nak()
			}
		}

		// Update pool metrics
		metrics.FlowWorkerPoolActive.WithLabelValues(w.flow.FlowID).Set(float64(w.pool.Running()))
		metrics.FlowWorkerPoolCapacity.WithLabelValues(w.flow.FlowID).Set(float64(w.pool.Cap()))
	}
}

// processBatch handles a batch of NATS messages:
// extract metadata → apply filter → apply column mapping → deep clone → write to sink → ACK/NAK.
func (w *FlowWorker) processBatch(ctx context.Context, msgs []jetstream.Msg) {
	batchStart := time.Now()
	events := make([]*domain.Event, 0, len(msgs))
	passedMsgs := make([]jetstream.Msg, 0, len(msgs))
	poolEvents := make([]*domain.Event, 0, len(msgs))

	for _, msg := range msgs {
		// Extract event metadata from NATS headers (zero-unmarshal pattern)
		ev := w.parseEventFromMsg(msg)

		// Apply filter — skip events that don't match
		if w.filter != nil && !w.filter.Evaluate(ev.Data) {
			// Event filtered out — ACK it (consumed but not written)
			if w.runtimeMetrics != nil {
				w.runtimeMetrics.RecordFlowFiltered(w.flow.FlowID, 1)
			}
			_ = msg.Ack()
			pool.PutEvent(ev)
			continue
		}

		// Apply column mapping to event data
		if len(w.mappings) > 0 && len(ev.Data) > 0 {
			mapped, err := ApplyColumnMappings(ev.Data, w.mappings)
			if err != nil {
				w.log.Warn("column mapping failed, moving event to DLQ",
					"err", err,
					"offset", ev.Offset)
				if w.runtimeMetrics != nil {
					w.runtimeMetrics.RecordFlowFailure(
						w.flow.FlowID,
						w.flow.SourceID,
						w.flow.SinkID,
						"mapping_error",
						err.Error(),
						1,
					)
				}
				if dlqErr := w.natsClient.MoveToDLQ(ctx, msg, ports.DLQMoveOptions{
					FlowID:     w.flow.FlowID,
					SinkID:     w.flow.SinkID,
					Reason:     fmt.Sprintf("mapping_error: %s", err.Error()),
					ErrorClass: cdcerrors.DLQErrorMapping,
				}); dlqErr != nil {
					w.log.Error("failed to move mapping error to DLQ", "err", dlqErr, "offset", ev.Offset)
					_ = msg.Nak()
				} else if w.runtimeMetrics != nil {
					w.runtimeMetrics.RecordDLQ(w.flow.FlowID, w.flow.SinkID, cdcerrors.DLQErrorMapping, 1)
				}
				pool.PutEvent(ev)
				continue
			}
			ev.Data = mapped
		}

		applySinkTable(ev, w.flow.SinkTable)
		events = append(events, ev)
		passedMsgs = append(passedMsgs, msg)
		poolEvents = append(poolEvents, ev)
	}

	if len(events) == 0 {
		return
	}

	// Write batch to sink with duration measurement
	start := time.Now()
	err := retry.Do(ctx, retry.DefaultConfig(), func() error {
		return w.sink.WriteBatch(events)
	})
	duration := time.Since(start)
	metrics.SinkWriteDuration.WithLabelValues(w.sink.InstanceID(), "").Observe(duration.Seconds())

	if err != nil {
		w.log.Error("sink write failed",
			"batch_size", len(events),
			"err", err)
		metrics.FlowEventsProcessed.WithLabelValues(w.flow.FlowID, "failure").Add(float64(len(events)))
		if w.runtimeMetrics != nil {
			w.runtimeMetrics.RecordFlowFailure(
				w.flow.FlowID,
				w.flow.SourceID,
				w.flow.SinkID,
				"sink_write_failed",
				err.Error(),
				uint64(len(events)),
			)
		}
		w.handleFailure(ctx, passedMsgs)
		// Return original events to pool
		for _, ev := range poolEvents {
			pool.PutEvent(ev)
		}
		return
	}

	// Record per-flow metrics
	metrics.FlowEventsProcessed.WithLabelValues(w.flow.FlowID, "success").Add(float64(len(events)))
	metrics.FlowBatchSize.WithLabelValues(w.flow.FlowID).Observe(float64(len(events)))
	metrics.FlowProcessingDuration.WithLabelValues(w.flow.FlowID).Observe(time.Since(batchStart).Seconds())

	if w.runtimeMetrics != nil {
		lastEvent := events[len(events)-1]
		w.runtimeMetrics.RecordSinkWrite(
			w.flow.FlowID,
			w.flow.SourceID,
			w.flow.SinkID,
			uint64(len(events)),
			duration.Milliseconds(),
			eventTimestampMs(lastEvent),
		)
	}

	// Success: ACK all messages
	for _, msg := range passedMsgs {
		_ = msg.Ack()
	}

	// Save offset checkpoint (use the last event's offset)
	lastEvent := events[len(events)-1]
	if lastEvent.Offset != "" {
		if err := w.store.SaveOffset(ctx, w.flow.FlowID, lastEvent.Offset); err != nil {
			w.log.Warn("failed to save offset",
				"offset", lastEvent.Offset,
				"err", err)
		}
	}

	w.log.Debug("batch processed",
		"count", len(events),
		"last_offset", lastEvent.Offset)

	// Return original events to pool
	for _, ev := range poolEvents {
		pool.PutEvent(ev)
	}
}

func applySinkTable(event *domain.Event, sinkTable string) {
	if event == nil || sinkTable == "" {
		return
	}
	schema, table := parseSourceTable(sinkTable)
	event.Schema = schema
	event.Table = table
}

// handleFailure handles batch write failures by NAKing messages or routing to DLQ
// if max retries have been exceeded.
func (w *FlowWorker) handleFailure(ctx context.Context, msgs []jetstream.Msg) {
	for _, msg := range msgs {
		// Check delivery count from message metadata
		metadata, err := msg.Metadata()
		if err != nil {
			// Can't determine delivery count — NAK for retry
			_ = msg.Nak()
			continue
		}

		if int(metadata.NumDelivered) >= w.maxDeliver {
			// Max retries exceeded — route to DLQ
			reason := fmt.Sprintf("max deliveries (%d) exceeded for flow %s", w.maxDeliver, w.flow.FlowID)
			if dlqErr := w.natsClient.MoveToDLQ(ctx, msg, ports.DLQMoveOptions{
				FlowID:     w.flow.FlowID,
				SinkID:     w.flow.SinkID,
				Reason:     reason,
				ErrorClass: cdcerrors.DLQErrorSink,
			}); dlqErr != nil {
				w.log.Error("failed to move message to DLQ",
					"err", dlqErr,
					"delivery_count", metadata.NumDelivered)
				// Last resort: NAK so it's not lost
				_ = msg.Nak()
			} else {
				metrics.DLQEventsTotal.WithLabelValues(w.flow.FlowID, "max_retries_exceeded").Inc()
				if w.runtimeMetrics != nil {
					w.runtimeMetrics.RecordDLQ(w.flow.FlowID, w.flow.SinkID, "max_retries_exceeded", 1)
				}
				w.log.Warn("message moved to DLQ",
					"delivery_count", metadata.NumDelivered,
					"subject", msg.Subject())
			}
		} else {
			// NAK for retry
			_ = msg.Nak()
		}
	}
}

func eventTimestampMs(ev *domain.Event) int64 {
	if ev == nil || ev.TimestampMS <= 0 {
		return time.Now().UnixMilli()
	}
	return ev.TimestampMS
}

func timestampMSFromData(data []byte) int64 {
	if len(data) == 0 {
		return 0
	}
	node, err := sonic.Get(data, "ts_ms")
	if err != nil || !node.Exists() {
		return 0
	}
	ts, err := node.Int64()
	if err != nil || ts <= 0 {
		return 0
	}
	return ts
}

// parseEventFromMsg extracts event metadata from NATS message headers.
// Uses the zero-unmarshal pattern: routing metadata is in headers,
// the raw payload stays as []byte without deserialization.
func (w *FlowWorker) parseEventFromMsg(msg jetstream.Msg) *domain.Event {
	ev := pool.GetEvent()
	headers := msg.Headers()

	ev.InstanceID = headers.Get(constant.HeaderInstanceID)
	ev.Offset = headers.Get(constant.HeaderOffset)
	ev.Schema = headers.Get(constant.HeaderSchema)
	ev.Table = headers.Get(constant.HeaderTable)
	ev.Op = constant.Op(headers.Get(constant.HeaderOp))
	ev.Data = msg.Data()

	if lsnStr := headers.Get(constant.HeaderLSN); lsnStr != "" {
		ev.LSN, _ = strconv.ParseUint(lsnStr, 10, 64)
	}
	ev.TimestampMS = timestampMSFromData(ev.Data)

	return ev
}

// Stop cancels the worker context, releases the ants pool via PoolManager,
// and waits for the run loop to exit.
func (w *FlowWorker) Stop() {
	w.cancel()
	w.poolManager.ReleasePool(w.flow.FlowID)
	<-w.stopped
	w.log.Info("worker stopped")
}
