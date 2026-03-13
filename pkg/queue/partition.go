package queue

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Partition struct {
	topic string
	id    int
	dir   string

	segments []*Segment
	active   *Segment

	maxSegmentSize int64
	indexInterval  int64
	nextOffset     atomic.Uint64
	mu             sync.RWMutex
}

func NewPartition(dataDir string, topic string, id int, maxSegmentSize int64, indexInterval int64) (*Partition, error) {
	dir := filepath.Join(dataDir, fmt.Sprintf("%s-%d", topic, id))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create partition dir: %w", err)
	}

	p := &Partition{
		topic:          topic,
		id:             id,
		dir:            dir,
		maxSegmentSize: maxSegmentSize,
		indexInterval:  indexInterval,
	}

	// Scan for existing segment files to recover state
	matches, _ := filepath.Glob(filepath.Join(dir, "*.log"))

	if len(matches) > 0 {
		sort.Strings(matches)

		for _, logPath := range matches {
			base := filepath.Base(logPath)
			var baseOffset uint64
			if n, _ := fmt.Sscanf(base, "%d.log", &baseOffset); n != 1 {
				continue
			}

			seg, err := OpenSegment(logPath, baseOffset, maxSegmentSize, indexInterval)
			if err != nil {
				slog.Error("failed to recover segment", "path", logPath, "err", err)
				continue
			}

			p.segments = append(p.segments, seg)
		}

		// Validate segment offset continuity
		for i := 0; i < len(p.segments)-1; i++ {
			curr := p.segments[i]
			next := p.segments[i+1]

			// Check for overlapping offset ranges
			if curr.baseOffset >= next.baseOffset {
				slog.Warn("segment offset overlap detected",
					"segment1_base", curr.baseOffset,
					"segment2_base", next.baseOffset)
			}
		}
	}

	if len(p.segments) > 0 {
		p.active = p.segments[len(p.segments)-1]

		// Restore nextOffset from highest written offset
		for _, seg := range p.segments {
			if seg.writePos.Load() > 0 {
				p.nextOffset.Store(seg.lastOffset.Load())
			}
		}
	} else {
		// Fresh partition — create first segment
		if err := p.createSegment(0); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *Partition) createSegment(baseOffset uint64) error {
	path := filepath.Join(p.dir, fmt.Sprintf("%020d.log", baseOffset))

	seg, err := OpenSegment(path, baseOffset, p.maxSegmentSize, p.indexInterval)
	if err != nil {
		return fmt.Errorf("failed to create segment: %w", err)
	}

	p.segments = append(p.segments, seg)
	p.active = seg
	return nil
}

func (p *Partition) rollSegment() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	base := p.nextOffset.Load() + 1
	return p.createSegment(base)
}

func (p *Partition) Produce(msg *Message) (uint64, error) {
	p.mu.RLock()
	active := p.active
	p.mu.RUnlock()

	offset := p.nextOffset.Add(1)
	msg.Offset = offset
	msg.Timestamp = time.Now().UnixNano()

	maxRetries := 3
	retries := 0

	for {
		_, _, err := active.Append(msg)
		if err == ErrSegmentFull {
			retries++
			if retries > maxRetries {
				return 0, fmt.Errorf("failed to roll segment after %d attempts: possible disk full", maxRetries)
			}

			if err := p.rollSegment(); err != nil {
				slog.Error("failed to roll segment", "err", err, "retries", retries)
				if retries >= maxRetries {
					return 0, fmt.Errorf("disk may be full: cannot create new segment")
				}
				// Backoff and retry
				time.Sleep(time.Duration(retries*100) * time.Millisecond)
				continue
			}
			p.mu.RLock()
			active = p.active
			p.mu.RUnlock()
			continue
		}
		if err != nil {
			return 0, err
		}
		return offset, nil
	}
}

func (p *Partition) Fetch(offset uint64, maxBytes int) ([]*MessageView, error) {
	p.mu.RLock()
	segments := make([]*Segment, len(p.segments))
	copy(segments, p.segments)
	p.mu.RUnlock()

	for _, seg := range segments {
		if offset < seg.baseOffset {
			continue
		}

		pos, err := seg.FindOffset(offset)
		if err != nil {
			return nil, err
		}

		msgs, _, err := seg.FetchBatch(pos, maxBytes)
		if err != nil && err != os.ErrNotExist {
			return nil, err
		}
		if len(msgs) > 0 {
			return msgs, nil
		}
		// No messages in this segment at this offset, try next
	}

	return nil, io.EOF
}

func (p *Partition) ProduceBatch(msgs []*Message) ([]uint64, error) {
	p.mu.RLock()
	active := p.active
	p.mu.RUnlock()

	for {
		offsets, err := active.AppendBatch(msgs)
		if err == ErrSegmentFull {
			if err := p.rollSegment(); err != nil {
				return nil, err
			}
			p.mu.RLock()
			active = p.active
			p.mu.RUnlock()
			continue
		}
		return offsets, err
	}
}

func (p *Partition) GetStats() QueueStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var size int64
	for _, s := range p.segments {
		size += s.Size()
	}
	return QueueStats{
		SegmentsCount: len(p.segments),
		TotalSizeMB:   size / (1024 * 1024),
		TotalEnqueued: p.nextOffset.Load(),
	}
}
