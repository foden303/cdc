package queue

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type MaintenanceMetrics struct {
	lastRetention      time.Time
	lastCompaction     time.Time
	lastFsync          time.Time
	compactionDuration atomic.Int64 // nanoseconds
	mu                 sync.Mutex
}

func NewMaintenanceMetrics() *MaintenanceMetrics {
	return &MaintenanceMetrics{
		lastRetention:  time.Now(),
		lastCompaction: time.Now(),
		lastFsync:      time.Now(),
	}
}

func (b *Broker) maintenanceLoop() {
	if b.metrics == nil {
		b.metrics = NewMaintenanceMetrics()
	}

	// Separate tickers for different maintenance tasks
	retentionTicker := time.NewTicker(5 * time.Minute)
	compactionTicker := time.NewTicker(30 * time.Second)
	fsyncTicker := time.NewTicker(10 * time.Second)

	defer retentionTicker.Stop()
	defer compactionTicker.Stop()
	defer fsyncTicker.Stop()

	for {
		select {
		case <-retentionTicker.C:
			b.runRetention()

		case <-compactionTicker.C:
			// Skip compaction if last one took too long (>5 seconds)
			lastDuration := time.Duration(b.metrics.compactionDuration.Load()) * time.Nanosecond
			if lastDuration > 5*time.Second {
				slog.Debug("skipping compaction, last run was too slow", "duration_ms", lastDuration.Milliseconds())
				break
			}
			b.runCompaction()

		case <-fsyncTicker.C:
			b.runFsync()

		case <-b.stop:
			return
		}
	}
}

func (b *Broker) runRetention() {
	b.mu.RLock()
	topics := make([]*Topic, 0, len(b.topics))
	for _, t := range b.topics {
		topics = append(topics, t)
	}
	b.mu.RUnlock()

	for _, t := range topics {
		for _, p := range t.partitions {
			p.Retention(RetentionPolicy{
				MaxSize: 1024 * 1024 * 1024, // 1GB default
			})
		}
	}

	b.metrics.mu.Lock()
	b.metrics.lastRetention = time.Now()
	b.metrics.mu.Unlock()

	slog.Debug("retention policy enforced")
}

func (b *Broker) runCompaction() {
	start := time.Now()

	b.mu.RLock()
	topics := make([]*Topic, 0, len(b.topics))
	for _, t := range b.topics {
		topics = append(topics, t)
	}
	b.mu.RUnlock()

	for _, t := range topics {
		for _, p := range t.partitions {
			p.compact()
		}
	}

	duration := time.Since(start)
	b.metrics.compactionDuration.Store(duration.Nanoseconds())
	b.metrics.mu.Lock()
	b.metrics.lastCompaction = time.Now()
	b.metrics.mu.Unlock()

	slog.Debug("compaction completed", "duration_ms", duration.Milliseconds())
}

func (b *Broker) runFsync() {
	b.mu.RLock()
	topics := make([]*Topic, 0, len(b.topics))
	for _, t := range b.topics {
		topics = append(topics, t)
	}
	b.mu.RUnlock()

	for _, t := range topics {
		for _, p := range t.partitions {
			p.mu.RLock()
			if p.active != nil {
				p.active.logFile.Sync()
				p.active.indexFile.Sync()
			}
			p.mu.RUnlock()
		}
	}

	b.metrics.mu.Lock()
	b.metrics.lastFsync = time.Now()
	b.metrics.mu.Unlock()
}

func (b *Broker) runMaintenance() {
	b.runRetention()
	b.runCompaction()
	b.runFsync()
}
