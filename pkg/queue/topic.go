package queue

import (
	"fmt"
	"sync/atomic"
)

type Topic struct {
	Name       string
	partitions []*Partition

	stickyPartition atomic.Uint32
	stickyCounter   atomic.Uint32
	indexInterval   int64
}

const stickyBatchSize = 1000

func NewTopic(dataDir string, name string, partitionCount int, maxSegmentSize int64, indexInterval int64) (*Topic, error) {
	t := &Topic{
		Name:          name,
		indexInterval: indexInterval,
	}

	for i := 0; i < partitionCount; i++ {
		p, err := NewPartition(dataDir, name, i, maxSegmentSize, indexInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to create partition %d: %w", i, err)
		}
		t.partitions = append(t.partitions, p)
	}

	t.stickyPartition.Store(0)

	return t, nil
}

func (t *Topic) Partition(key []byte) *Partition {
	if key == nil {
		// Sticky Partitioner logic
		count := t.stickyCounter.Add(1)
		curr := t.stickyPartition.Load()

		if count >= stickyBatchSize {
			if t.stickyCounter.CompareAndSwap(count, 0) {
				next := (curr + 1) % uint32(len(t.partitions))
				t.stickyPartition.Store(next)
				return t.partitions[next]
			}
		}
		return t.partitions[curr]
	}

	hash := uint32(0)
	for _, b := range key {
		hash = hash*31 + uint32(b)
	}

	return t.partitions[hash%uint32(len(t.partitions))]
}

func (t *Topic) PartitionsCount() int {
	return len(t.partitions)
}

func (t *Topic) PartitionByID(id int) *Partition {
	if id < 0 || id >= len(t.partitions) {
		return nil
	}
	return t.partitions[id]
}
