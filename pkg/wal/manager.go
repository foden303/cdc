package wal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/foden/cdc/pkg/cluster"
	"github.com/foden/cdc/pkg/queue"
)

type WalMessage[T any] struct {
	Offset    uint64
	Timestamp time.Time
	Key       []byte
	Item      T
}

type Manager[T any] struct {
	mu sync.RWMutex

	dir            string
	topic          string
	maxSegmentSize int64
	retentionHours int

	broker   *queue.Broker
	raftNode *cluster.RaftNode

	// Per-partition consumed offsets for DequeueBatch
	consumedOffsets map[int]uint64
}

// OpenManager initializes a new WAL manager.
func OpenManager[T any](dir string, topicName string, partitionCount int, maxSegmentSize int64, retentionHours int) (*Manager[T], error) {

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create manager dir: %w", err)
	}

	broker := queue.NewBroker(dir, maxSegmentSize)

	if err := broker.CreateTopic(topicName, partitionCount); err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}

	m := &Manager[T]{
		dir:             dir,
		topic:           topicName,
		maxSegmentSize:  maxSegmentSize,
		retentionHours:  retentionHours,
		broker:          broker,
		consumedOffsets: make(map[int]uint64),
	}

	return m, nil
}

func (m *Manager[T]) GetBroker() *queue.Broker {
	return m.broker
}

// SetRaftNode sets the Raft node for cluster mode.
// When set, Enqueue routes writes through Raft consensus.
func (m *Manager[T]) SetRaftNode(node *cluster.RaftNode) {
	m.raftNode = node
}

func (m *Manager[T]) GetTopic() *queue.Topic {
	return m.broker.Topic(m.topic)
}

func (m *Manager[T]) GetPartition(id int) *queue.Partition {
	topic := m.GetTopic()
	if topic == nil {
		return nil
	}
	return topic.PartitionByID(id)
}

func (m *Manager[T]) Enqueue(key []byte, item T) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	// Cluster mode: route through Raft consensus
	if m.raftNode != nil {
		cmd := cluster.ProduceCommand{
			Topic:     m.topic,
			Key:       key,
			Value:     data,
			Timestamp: time.Now().UnixNano(),
		}
		return m.raftNode.Propose(cluster.CmdProduce, cmd)
	}

	// Standalone mode: direct produce
	msg := &queue.Message{
		Key:   key,
		Value: data,
	}

	_, err = m.GetTopic().Partition(key).Produce(msg)
	return err
}

func (m *Manager[T]) DequeueBatch(partitionID int, batchSize int) ([]WalMessage[T], error) {
	part := m.GetPartition(partitionID)
	if part == nil {
		return nil, fmt.Errorf("partition %d not found", partitionID)
	}

	m.mu.RLock()
	offset := m.consumedOffsets[partitionID]
	m.mu.RUnlock()

	msgs, err := part.Fetch(offset, batchSize*1024)
	if err != nil {
		if err == io.EOF {
			return nil, nil // no new messages
		}
		return nil, err
	}

	items := make([]WalMessage[T], 0, len(msgs))
	var maxOffset uint64

	for _, msg := range msgs {
		var decoded T
		if err := json.Unmarshal(msg.Value, &decoded); err != nil {
			continue
		}

		items = append(items, WalMessage[T]{
			Offset:    msg.Offset,
			Timestamp: time.Unix(0, msg.Timestamp),
			Key:       msg.Key,
			Item:      decoded,
		})

		if msg.Offset >= maxOffset {
			maxOffset = msg.Offset + 1
		}
	}

	// Advance consumed offset
	if len(items) > 0 {
		m.mu.Lock()
		m.consumedOffsets[partitionID] = maxOffset
		m.mu.Unlock()
	}

	return items, nil
}

func (m *Manager[T]) Commit(partitionID int) error {
	// Offset is already tracked in consumedOffsets via DequeueBatch
	return nil
}

func (m *Manager[T]) Close() error {
	if m.raftNode != nil {
		if err := m.raftNode.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown raft: %w", err)
		}
	}
	m.broker.Close()
	return nil
}

func (m *Manager[T]) GetPartitionIDs() []int {
	topic := m.GetTopic()
	if topic == nil {
		return nil
	}
	var ids []int
	for i := 0; i < topic.PartitionsCount(); i++ {
		ids = append(ids, i)
	}
	return ids
}

func (m *Manager[T]) InspectRaw(partitionID int, startOffset uint64, limit int) ([]WalMessage[T], error) {
	part := m.GetPartition(partitionID)
	if part == nil {
		return nil, fmt.Errorf("partition %d not found", partitionID)
	}
	msgs, err := part.Fetch(startOffset, limit*1024)
	if err != nil && err != io.EOF {
		return nil, err
	}

	items := make([]WalMessage[T], 0, len(msgs))
	for _, msg := range msgs {
		var decoded T
		if err := json.Unmarshal(msg.Value, &decoded); err == nil {
			items = append(items, WalMessage[T]{
				Offset:    msg.Offset,
				Timestamp: time.Unix(0, msg.Timestamp),
				Key:       msg.Key,
				Item:      decoded,
			})
		}
	}
	return items, nil
}

func (m *Manager[T]) GetTotalStats() queue.QueueStats {
	topic := m.GetTopic()
	if topic == nil {
		return queue.QueueStats{}
	}

	var total queue.QueueStats
	for i := 0; i < topic.PartitionsCount(); i++ {
		p := topic.PartitionByID(i)
		if p == nil {
			continue
		}
		stats := p.GetStats()
		total.TotalEnqueued += stats.TotalEnqueued
		total.TotalSizeMB += stats.TotalSizeMB
		total.SegmentsCount += stats.SegmentsCount
	}
	return total
}
