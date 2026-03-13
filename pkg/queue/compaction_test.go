package queue

import (
	"os"
	"testing"
	"time"
)

func TestLogCompaction(t *testing.T) {
	testDir := t.TempDir()

	p, err := NewPartition(testDir, "compact_test", 0, 1024*1024*1024, 4096)
	if err != nil {
		t.Fatalf("failed to create partition: %v", err)
	}

	// Write messages with same key to first segment
	p.Produce(&Message{Key: []byte("k1"), Value: []byte("v1"), Timestamp: time.Now().Unix()})
	p.Produce(&Message{Key: []byte("k1"), Value: []byte("v2"), Timestamp: time.Now().Unix()})
	p.Produce(&Message{Key: []byte("k2"), Value: []byte("v1"), Timestamp: time.Now().Unix()})

	// Roll segment
	p.rollSegment()

	// Write more messages to second segment (active)
	p.Produce(&Message{Key: []byte("k1"), Value: []byte("v3"), Timestamp: time.Now().Unix()})
	p.Produce(&Message{Key: []byte("k2"), Value: []byte("v2"), Timestamp: time.Now().Unix()})

	// Roll again so first and second are eligible for compaction
	p.rollSegment()

	// Initial segments check
	p.mu.RLock()
	initialCount := len(p.segments)
	p.mu.RUnlock()
	if initialCount != 3 {
		t.Fatalf("expected 3 segments, got %d", initialCount)
	}

	// Compact
	if err := p.compact(); err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	// Should now have 2 segments (1 compacted + 1 active)
	p.mu.RLock()
	finalCount := len(p.segments)
	p.mu.RUnlock()
	if finalCount != 2 {
		t.Fatalf("expected 2 segments after compaction, got %d", finalCount)
	}

	// Verify latest values only
	msgs, err := p.Fetch(1, 1024*1024)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	foundK1 := false
	foundK2 := false
	for _, m := range msgs {
		if string(m.Key) == "k1" {
			if string(m.Value) != "v3" {
				t.Errorf("expected k1=v3, got %s", string(m.Value))
			}
			foundK1 = true
		}
		if string(m.Key) == "k2" {
			if string(m.Value) != "v2" {
				t.Errorf("expected k2=v2, got %s", string(m.Value))
			}
			foundK2 = true
		}
	}

	if !foundK1 || !foundK2 {
		t.Errorf("missing keys after compaction: k1=%v, k2=%v", foundK1, foundK2)
	}

	// Cleanup
	os.RemoveAll(testDir)
}
