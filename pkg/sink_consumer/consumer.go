package sink_consumer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/foden/cdc/pkg/interfaces"
	"github.com/foden/cdc/pkg/queue"
)

// SinkConsumer consumes messages from a queue topic and writes them to multiple sinks
type SinkConsumer struct {
	// Configuration (exported for API responses)
	ID            string
	Name          string
	topicName     string
	partitionIDs  []int // Subscribed partitions (nil = all)
	sinks         []interfaces.Sink
	batchSize     int
	flushInterval time.Duration
	handler       MessageHandler

	// Queue integration
	broker      *queue.Broker
	coordinator *queue.Coordinator
	group       string // Consumer group ID
	partitions  map[int]*queue.Partition

	// State management
	offsets map[int]uint64 // partition_id -> current offset
	mu      sync.RWMutex

	// Lifecycle
	stopCh    chan struct{}
	stoppedCh chan struct{}
	metrics   *ConsumerMetrics

	// Worker control
	workerWg sync.WaitGroup
	paused   bool
}

// NewSinkConsumer creates a new sink consumer
func NewSinkConsumer(
	id string,
	name string,
	topicName string,
	sinks []interfaces.Sink,
	broker *queue.Broker,
	coordinator *queue.Coordinator,
	config CreateConsumerConfig,
) (*SinkConsumer, error) {
	if id == "" {
		return nil, fmt.Errorf("consumer id cannot be empty")
	}
	if name == "" {
		name = id
	}
	if topicName == "" {
		return nil, fmt.Errorf("topic name cannot be empty")
	}
	if len(sinks) == 0 {
		return nil, fmt.Errorf("at least one sink is required")
	}

	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 500
	}

	flushIntervalMs := config.FlushIntervalMs
	if flushIntervalMs <= 0 {
		flushIntervalMs = 1000
	}

	sc := &SinkConsumer{
		ID:            id,
		Name:          name,
		topicName:     topicName,
		partitionIDs:  config.Partitions,
		sinks:         sinks,
		batchSize:     batchSize,
		flushInterval: time.Duration(flushIntervalMs) * time.Millisecond,
		handler:       NewJSONMessageHandler(),
		broker:        broker,
		coordinator:   coordinator,
		group:         fmt.Sprintf("sink-group-%s", id),
		offsets:       make(map[int]uint64),
		stopCh:        make(chan struct{}),
		stoppedCh:     make(chan struct{}),
		partitions:    make(map[int]*queue.Partition),
		metrics: &ConsumerMetrics{
			StartTime: time.Now(),
		},
	}

	return sc, nil
}

// Start begins consuming from the topic
func (sc *SinkConsumer) Start(ctx context.Context) error {
	// Get topic from broker
	topic := sc.broker.Topic(sc.topicName)
	if topic == nil {
		return fmt.Errorf("topic %s not found", sc.topicName)
	}

	// Determine partitions to subscribe to
	partitionCount := topic.PartitionsCount()
	var targetPartitions []int
	if len(sc.partitionIDs) > 0 {
		targetPartitions = sc.partitionIDs
	} else {
		// Subscribe to all partitions
		targetPartitions = make([]int, partitionCount)
		for i := 0; i < partitionCount; i++ {
			targetPartitions[i] = i
		}
	}

	// Register with coordinator and get partition assignment
	sc.mu.Lock()
	for _, partID := range targetPartitions {
		partition := topic.PartitionByID(partID)
		if partition == nil {
			sc.mu.Unlock()
			return fmt.Errorf("partition %d not found in topic %s", partID, sc.topicName)
		}
		sc.partitions[partID] = partition

		// Load last committed offset for this partition
		offset, err := sc.coordinator.FetchOffset(sc.group, sc.topicName, partID)
		if err != nil {
			// If offset not found (group doesn't exist yet), start from 0
			offset = 0
		}
		sc.offsets[partID] = offset
	}
	sc.mu.Unlock()

	slog.Info("sink consumer starting",
		"consumer_id", sc.ID,
		"topic", sc.topicName,
		"partitions", len(sc.partitions))

	// Start metrics goroutine
	sc.workerWg.Add(1)
	go sc.metricsLoop()

	// Start partition workers
	for partID, partition := range sc.partitions {
		sc.workerWg.Add(1)
		go sc.partitionWorker(partID, partition)
	}

	return nil
}

// Stop gracefully shuts down the consumer
func (sc *SinkConsumer) Stop() error {
	slog.Info("stopping sink consumer", "consumer_id", sc.ID)
	close(sc.stopCh)

	// Wait for all workers and metrics goroutine to finish
	done := make(chan struct{})
	go func() {
		sc.workerWg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		slog.Info("sink consumer stopped", "consumer_id", sc.ID)
	case <-time.After(30 * time.Second):
		slog.Warn("sink consumer shutdown timeout exceeded", "consumer_id", sc.ID)
	}

	// Close all sinks
	for _, sink := range sc.sinks {
		if err := sink.Close(); err != nil {
			slog.Error("failed to close sink", "consumer_id", sc.ID, "err", err)
		}
	}

	close(sc.stoppedCh)
	return nil
}

// Pause pauses message consumption (but keeps connections alive)
func (sc *SinkConsumer) Pause() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.paused = true
	slog.Info("sink consumer paused", "consumer_id", sc.ID)
}

// Resume resumes message consumption
func (sc *SinkConsumer) Resume() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.paused = false
	slog.Info("sink consumer resumed", "consumer_id", sc.ID)
}

// GetStats returns current consumer statistics
func (sc *SinkConsumer) GetStats() *ConsumerStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	lastOffset := make(map[int]uint64)
	for partID, offset := range sc.offsets {
		lastOffset[partID] = offset
	}

	status := "running"
	if sc.paused {
		status = "paused"
	}

	sinkNames := make([]string, len(sc.sinks))
	for i, sink := range sc.sinks {
		sinkNames[i] = fmt.Sprintf("%T", sink)
	}

	upSince := sc.metrics.StartTime

	return &ConsumerStats{
		ID:             sc.ID,
		Name:           sc.Name,
		TopicName:      sc.topicName,
		Status:         status,
		TotalProcessed: sc.metrics.TotalProcessed,
		TotalErrors:    sc.metrics.TotalErrors,
		LastOffset:     lastOffset,
		LastUpdated:    time.Now(),
		UpSince:        &upSince,
		Sinks:          sinkNames,
	}
}

// partitionWorker processes messages from a single partition
func (sc *SinkConsumer) partitionWorker(partID int, partition *queue.Partition) {
	defer sc.workerWg.Done()

	slog.Info("partition worker started",
		"consumer_id", sc.ID,
		"partition_id", partID)

	// Idle sleep duration
	idleSleep := 100 * time.Millisecond
	flushTicker := time.NewTicker(sc.flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case <-sc.stopCh:
			// Flush any remaining messages
			if err := sc.flushAllSinks(); err != nil {
				slog.Error("failed to flush on shutdown",
					"consumer_id", sc.ID,
					"partition", partID,
					"err", err)
			}
			slog.Info("partition worker stopped",
				"consumer_id", sc.ID,
				"partition_id", partID)
			return

		case <-flushTicker.C:
			// Periodic flush
			if err := sc.flushAllSinks(); err != nil {
				slog.Error("failed to periodic flush",
					"consumer_id", sc.ID,
					"partition", partID,
					"err", err)
			}

		default:
			// Check if paused
			sc.mu.RLock()
			paused := sc.paused
			offset := sc.offsets[partID]
			sc.mu.RUnlock()

			if paused {
				time.Sleep(idleSleep)
				continue
			}

			// Fetch messages
			messages, err := partition.Fetch(offset, 1<<20) // 1MB
			if err != nil {
				if err == io.EOF {
					// No more messages
					time.Sleep(idleSleep)
					continue
				}
				slog.Error("failed to fetch messages",
					"consumer_id", sc.ID,
					"partition", partID,
					"err", err)
				time.Sleep(idleSleep)
				continue
			}

			// Process messages
			processed := 0
			errors := 0

			for _, msg := range messages {
				// Deserialize message
				event, err := sc.handler.Handle(msg)
				if err != nil {
					slog.Error("message deserialization failed",
						"consumer_id", sc.ID,
						"partition", partID,
						"offset", msg.Offset,
						"err", err)
					errors++
					continue
				}

				// Write to all sinks
				writeErr := false
				for _, sink := range sc.sinks {
					if err := sink.Write(event); err != nil {
						slog.Error("sink write failed",
							"consumer_id", sc.ID,
							"partition", partID,
							"offset", msg.Offset,
							"sink", fmt.Sprintf("%T", sink),
							"err", err)
						errors++
						writeErr = true
						break
					}
				}

				if !writeErr {
					processed++
				}

				// Update offset
				sc.mu.Lock()
				sc.offsets[partID] = msg.Offset + 1
				sc.metrics.TotalProcessed++
				sc.metrics.LastUpdate = time.Now()
				sc.mu.Unlock()

				// Check if batch size reached
				if processed >= sc.batchSize {
					if err := sc.flushAllSinks(); err != nil {
						slog.Error("failed to flush after batch",
							"consumer_id", sc.ID,
							"partition", partID,
							"err", err)
						errors++
					}
					// Commit offset
					sc.commitOffset(partID)
					processed = 0
				}
			}

			// Update error count
			if errors > 0 {
				sc.mu.Lock()
				sc.metrics.TotalErrors += uint64(errors)
				sc.mu.Unlock()
			}
		}
	}
}

// flushAllSinks flushes all registered sinks
func (sc *SinkConsumer) flushAllSinks() error {
	for _, sink := range sc.sinks {
		if err := sink.Flush(); err != nil {
			return fmt.Errorf("failed to flush sink: %w", err)
		}
	}
	return nil
}

// commitOffset persists the partition offset to the coordinator
func (sc *SinkConsumer) commitOffset(partID int) {
	sc.mu.RLock()
	offset := sc.offsets[partID]
	sc.mu.RUnlock()

	sc.coordinator.CommitOffset(sc.group, sc.topicName, partID, offset)
}

// metricsLoop periodically logs consumer metrics
func (sc *SinkConsumer) metricsLoop() {
	defer sc.workerWg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sc.stopCh:
			return
		case <-ticker.C:
			stats := sc.GetStats()
			slog.Info("sink consumer metrics",
				"consumer_id", stats.ID,
				"status", stats.Status,
				"total_processed", stats.TotalProcessed,
				"total_errors", stats.TotalErrors,
				"partitions", len(stats.LastOffset))
		}
	}
}
