package runtime

import (
	"sync"
	"sync/atomic"

	"github.com/foden/cdc/internal/core/ports"
)

type TableInterestSnapshot struct {
	PartitionCount int32
	FlowCount      int32
	Active         bool
}

type Registry struct {
	flows  sync.Map
	tables sync.Map
}

type tableEntry struct {
	mu       sync.Mutex
	flows    map[string]*FlowRuntimeInfo
	snapshot atomic.Pointer[TableInterestSnapshot]
}

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{}
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func SetDefaultRegistry(registry *Registry) {
	if registry != nil {
		defaultRegistry = registry
	}
}

func (r *Registry) RegisterFlow(flow *ports.FlowConfig) error {
	info, err := NewFlowRuntimeInfo(flow)
	if err != nil {
		return err
	}

	if oldAny, loaded := r.flows.LoadOrStore(info.FlowID, info); loaded {
		old := oldAny.(*FlowRuntimeInfo)
		r.removeTableInterest(old)
		r.flows.Store(info.FlowID, info)
	}

	r.addTableInterest(info)
	return nil
}

func (r *Registry) UnregisterFlow(flowID string) {
	if flowID == "" {
		return
	}
	if existing, ok := r.flows.LoadAndDelete(flowID); ok {
		r.removeTableInterest(existing.(*FlowRuntimeInfo))
	}
}

func (r *Registry) LookupFlow(flowID string) (*FlowRuntimeInfo, bool) {
	existing, ok := r.flows.Load(flowID)
	if !ok {
		return nil, false
	}
	return existing.(*FlowRuntimeInfo), true
}

func (r *Registry) LookupTable(sourceID, schema, table string) (TableInterestSnapshot, bool) {
	existing, ok := r.tables.Load(tableKey{sourceID: sourceID, schema: schema, table: table})
	if !ok {
		return TableInterestSnapshot{}, false
	}
	snapshot := existing.(*tableEntry).snapshot.Load()
	if snapshot == nil || !snapshot.Active {
		return TableInterestSnapshot{}, false
	}
	return *snapshot, true
}

func (r *Registry) addTableInterest(info *FlowRuntimeInfo) {
	key := newTableKey(info.SourceID, info.SourceTable)
	newEntry := &tableEntry{flows: make(map[string]*FlowRuntimeInfo)}
	actual, _ := r.tables.LoadOrStore(key, newEntry)
	entry := actual.(*tableEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()
	if entry.flows == nil {
		entry.flows = make(map[string]*FlowRuntimeInfo)
	}
	entry.flows[info.FlowID] = info
	entry.rebuildSnapshotLocked()
}

func (r *Registry) removeTableInterest(info *FlowRuntimeInfo) {
	key := newTableKey(info.SourceID, info.SourceTable)
	existing, ok := r.tables.Load(key)
	if !ok {
		return
	}

	entry := existing.(*tableEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	delete(entry.flows, info.FlowID)
	if len(entry.flows) == 0 {
		entry.snapshot.Store(&TableInterestSnapshot{})
		r.tables.Delete(key)
		return
	}
	entry.rebuildSnapshotLocked()
}

func (e *tableEntry) rebuildSnapshotLocked() {
	var maxPartition int32
	for _, flow := range e.flows {
		if flow.PartitionCount > maxPartition {
			maxPartition = flow.PartitionCount
		}
	}
	e.snapshot.Store(&TableInterestSnapshot{
		PartitionCount: maxPartition,
		FlowCount:      int32(len(e.flows)),
		Active:         len(e.flows) > 0,
	})
}
