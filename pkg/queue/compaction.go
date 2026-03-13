package queue

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func (p *Partition) compact() error {
	// Simple key-based compaction
	// 1. Collect all non-active segments
	p.mu.Lock()
	if len(p.segments) <= 1 {
		p.mu.Unlock()
		return nil
	}

	toCompact := make([]*Segment, len(p.segments)-1)
	copy(toCompact, p.segments[:len(p.segments)-1])
	p.mu.Unlock()

	latestOffsets := make(map[string]uint64)

	// Pass 1: Find latest offset for each key
	for _, seg := range toCompact {
		pos := int64(0)
		for {
			msg, next, err := seg.ReadAt(pos)
			if err != nil {
				if err == io.EOF || err == os.ErrNotExist {
					break
				}
				return err
			}
			latestOffsets[string(msg.Key)] = msg.Offset
			pos = next
		}
	}

	// Pass 2: Write latest messages to a new compacted segment
	compactedPath := filepath.Join(p.dir, fmt.Sprintf("%020d.log", toCompact[0].baseOffset))
	tmpPath := compactedPath + ".compacting"

	newSeg, err := OpenSegment(tmpPath, toCompact[0].baseOffset, p.maxSegmentSize, p.indexInterval)
	if err != nil {
		return err
	}

	for _, seg := range toCompact {
		pos := int64(0)
		for {
			msg, next, err := seg.ReadAt(pos)
			if err != nil {
				if err == io.EOF || err == os.ErrNotExist {
					break
				}
				newSeg.Close()
				return err
			}

			if latestOffsets[string(msg.Key)] == msg.Offset {
				slog.Debug("compacting", "key", string(msg.Key), "offset", msg.Offset)
				m := &Message{
					Offset:    msg.Offset,
					Key:       msg.Key,
					Value:     msg.Value,
					Timestamp: msg.Timestamp,
				}
				newSeg.Append(m)
			}
			pos = next
		}
	}

	// 3. Atomically replace segments
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close old segments and remove files
	for _, seg := range toCompact {
		seg.Close()
		os.Remove(seg.logFile.Name())
		os.Remove(seg.indexFile.Name())
	}

	// Rename compacted file to final name
	os.Rename(tmpPath, compactedPath)
	idxTmp := tmpPath[:len(tmpPath)-len(filepath.Ext(tmpPath))] + ".index"
	idxFinal := compactedPath[:len(compactedPath)-len(filepath.Ext(compactedPath))] + ".index"
	os.Rename(idxTmp, idxFinal)

	// Reopen the renamed segment
	newSeg.Close()
	finalSeg, err := OpenSegment(compactedPath, toCompact[0].baseOffset, p.maxSegmentSize, p.indexInterval)
	if err != nil {
		return err
	}

	newSegments := []*Segment{finalSeg}
	newSegments = append(newSegments, p.segments[len(toCompact):]...)
	p.segments = newSegments

	return nil
}
