package utils

import (
	"testing"

	"pgregory.net/rapid"
)

// TestProperty_PartitionDeterministic verifies that the same key and partitionCount
// always produces the same partition_id.
// **Validates: Requirements 7.3**
func TestProperty_PartitionDeterministic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		key := rapid.String().Draw(t, "key")
		count := rapid.IntRange(1, 100).Draw(t, "count")
		p1 := GeneratePartition(key, count)
		p2 := GeneratePartition(key, count)
		if p1 != p2 {
			t.Fatalf("non-deterministic: %d != %d for key=%q count=%d", p1, p2, key, count)
		}
	})
}

// TestProperty_PartitionBounded verifies that the result is always in range [0, partitionCount).
// **Validates: Requirements 7.3**
func TestProperty_PartitionBounded(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		key := rapid.String().Draw(t, "key")
		count := rapid.IntRange(1, 1000).Draw(t, "count")
		p := GeneratePartition(key, count)
		if p < 0 || p >= count {
			t.Fatalf("out of bounds: %d not in [0,%d) for key=%q", p, count, key)
		}
	})
}

// TestProperty_PartitionDistribution verifies that different random keys distribute
// across multiple partitions (not all going to partition 0).
// **Validates: Requirements 7.3**
func TestProperty_PartitionDistribution(t *testing.T) {
	const partitionCount = 10
	const numKeys = 200

	rapid.Check(t, func(t *rapid.T) {
		keys := rapid.SliceOfN(rapid.String(), numKeys, numKeys).Draw(t, "keys")
		seen := make(map[int]bool)
		for _, k := range keys {
			p := GeneratePartition(k, partitionCount)
			seen[p] = true
		}
		// With 200 random keys and 10 partitions, we should see at least 2 different partitions
		if len(seen) < 2 {
			t.Fatalf("poor distribution: only %d partitions used out of %d with %d keys", len(seen), partitionCount, numKeys)
		}
	})
}
