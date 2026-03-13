package queue

import (
	"fmt"
	"sync"
)

type Broker struct {
	dataDir        string
	maxSegmentSize int64
	indexInterval  int64
	topics         map[string]*Topic
	mu             sync.RWMutex
	stop           chan struct{}
	metrics        *MaintenanceMetrics
}

func NewBroker(dataDir string, maxSegmentSize int64) *Broker {
	b := &Broker{
		dataDir:        dataDir,
		maxSegmentSize: maxSegmentSize,
		indexInterval:  4096, // default 4KB
		topics:         make(map[string]*Topic),
		stop:           make(chan struct{}),
	}
	go b.maintenanceLoop()
	return b
}

// SetIndexInterval sets the index interval for all new segments
func (b *Broker) SetIndexInterval(interval int64) {
	b.indexInterval = interval
}

func (b *Broker) CreateTopic(name string, partitions int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.topics[name]; ok {
		return nil // topic already exists
	}

	t, err := NewTopic(b.dataDir, name, partitions, b.maxSegmentSize, b.indexInterval)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", name, err)
	}

	b.topics[name] = t
	return nil
}

func (b *Broker) Topic(name string) *Topic {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.topics[name]
}

func (b *Broker) Close() {
	close(b.stop)
}

type ConsumerGroup struct {
	ID string

	consumers map[string]*Consumer

	assignments map[string]map[int]string
	offsets     map[string]map[int]uint64
	mu          sync.RWMutex
}
