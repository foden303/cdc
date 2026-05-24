package registry

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// TableEntry holds metadata for a registered table.
type TableEntry struct {
	PartitionCount atomic.Int32
	RefCount       atomic.Int32
}

// TableRegistry tracks which source tables have active flows.
// Uses sync.Map for lock-free reads on the hot path (WAL event processing).
type TableRegistry struct {
	// key: "source_id:schema.table" → *TableEntry
	tables sync.Map
}

// NewTableRegistry creates a new TableRegistry.
func NewTableRegistry() *TableRegistry {
	return &TableRegistry{}
}

// buildKey constructs the lookup key for the registry.
func buildKey(sourceID, schema, table string) string {
	return fmt.Sprintf("%s:%s.%s", sourceID, schema, table)
}

// Register adds or updates a table entry. If already registered,
// takes max partition count and increments refCount.
func (r *TableRegistry) Register(sourceID, schema, table string, partitionCount int) {
	key := buildKey(sourceID, schema, table)

	// Try to load existing entry first (fast path)
	if existing, ok := r.tables.Load(key); ok {
		entry := existing.(*TableEntry)
		entry.RefCount.Add(1)
		// CAS loop to update partition count to max
		for {
			current := entry.PartitionCount.Load()
			if int32(partitionCount) <= current {
				break
			}
			if entry.PartitionCount.CompareAndSwap(current, int32(partitionCount)) {
				break
			}
		}
		return
	}

	// Slow path: create new entry
	newEntry := &TableEntry{}
	newEntry.PartitionCount.Store(int32(partitionCount))
	newEntry.RefCount.Store(1)

	if actual, loaded := r.tables.LoadOrStore(key, newEntry); loaded {
		// Another goroutine stored first — update the existing entry
		entry := actual.(*TableEntry)
		entry.RefCount.Add(1)
		for {
			current := entry.PartitionCount.Load()
			if int32(partitionCount) <= current {
				break
			}
			if entry.PartitionCount.CompareAndSwap(current, int32(partitionCount)) {
				break
			}
		}
	}
}

// Unregister decrements refCount. Removes entry if refCount reaches 0.
func (r *TableRegistry) Unregister(sourceID, schema, table string) {
	key := buildKey(sourceID, schema, table)
	if existing, ok := r.tables.Load(key); ok {
		entry := existing.(*TableEntry)
		if entry.RefCount.Add(-1) <= 0 {
			r.tables.Delete(key)
		}
	}
}

// Lookup returns the partition count for a table, or 0 if not registered.
// This is the hot-path method called on every WAL event — must be O(1) lock-free.
func (r *TableRegistry) Lookup(sourceID, schema, table string) (partitionCount int, active bool) {
	key := buildKey(sourceID, schema, table)
	if existing, ok := r.tables.Load(key); ok {
		entry := existing.(*TableEntry)
		return int(entry.PartitionCount.Load()), true
	}
	return 0, false
}

// GlobalTableRegistry is the shared singleton used by sources and the flow manager.
var GlobalTableRegistry = NewTableRegistry()
