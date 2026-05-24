package flow

import (
	"fmt"
	"sync"

	"github.com/panjf2000/ants/v2"
)

// PoolHealth represents the health status of a flow's worker pool.
type PoolHealth string

const (
	// PoolHealthIdle indicates the pool has less than 10% utilization.
	PoolHealthIdle PoolHealth = "idle"
	// PoolHealthActive indicates the pool has 10-80% utilization.
	PoolHealthActive PoolHealth = "active"
	// PoolHealthSaturated indicates the pool has more than 80% utilization.
	PoolHealthSaturated PoolHealth = "saturated"
)

// PoolMetrics holds runtime metrics for a single flow's pool.
type PoolMetrics struct {
	FlowID         string     `json:"flow_id"`
	RunningWorkers int        `json:"running_workers"`
	WaitingTasks   int        `json:"waiting_tasks"`
	PoolCapacity   int        `json:"pool_capacity"`
	Health         PoolHealth `json:"health"`
}

// PoolManager manages the lifecycle of all flow worker pools.
type PoolManager struct {
	mu    sync.RWMutex
	pools map[string]*ants.Pool
}

// NewPoolManager creates a new PoolManager instance.
func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools: make(map[string]*ants.Pool),
	}
}

// CreatePool creates a dedicated ants.Pool for a flow with the given size.
func (pm *PoolManager) CreatePool(flowID string, size int) (*ants.Pool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("pool size must be positive, got %d", size)
	}

	pool, err := ants.NewPool(size)
	if err != nil {
		return nil, err
	}

	pm.mu.Lock()
	pm.pools[flowID] = pool
	pm.mu.Unlock()

	return pool, nil
}

// ReleasePool releases a flow's pool and removes it from the manager.
func (pm *PoolManager) ReleasePool(flowID string) {
	pm.mu.Lock()
	pool, ok := pm.pools[flowID]
	if ok {
		delete(pm.pools, flowID)
	}
	pm.mu.Unlock()

	if ok {
		pool.Release()
	}
}

// GetMetrics returns metrics for a specific flow's pool.
// Returns nil if the flow has no active pool.
func (pm *PoolManager) GetMetrics(flowID string) *PoolMetrics {
	pm.mu.RLock()
	pool, ok := pm.pools[flowID]
	pm.mu.RUnlock()

	if !ok {
		return nil
	}

	return &PoolMetrics{
		FlowID:         flowID,
		RunningWorkers: pool.Running(),
		WaitingTasks:   pool.Waiting(),
		PoolCapacity:   pool.Cap(),
		Health:         calculateHealth(pool),
	}
}

// GetAllMetrics returns metrics for all active pools.
func (pm *PoolManager) GetAllMetrics() []*PoolMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	metrics := make([]*PoolMetrics, 0, len(pm.pools))
	for flowID, pool := range pm.pools {
		metrics = append(metrics, &PoolMetrics{
			FlowID:         flowID,
			RunningWorkers: pool.Running(),
			WaitingTasks:   pool.Waiting(),
			PoolCapacity:   pool.Cap(),
			Health:         calculateHealth(pool),
		})
	}
	return metrics
}

// GetHealth returns the health status of a flow's pool.
// Returns PoolHealthIdle if the flow has no active pool.
func (pm *PoolManager) GetHealth(flowID string) PoolHealth {
	pm.mu.RLock()
	pool, ok := pm.pools[flowID]
	pm.mu.RUnlock()

	if !ok {
		return PoolHealthIdle
	}

	return calculateHealth(pool)
}

// PauseAll releases all pools managed by the PoolManager.
func (pm *PoolManager) PauseAll() {
	pm.mu.Lock()
	pools := pm.pools
	pm.pools = make(map[string]*ants.Pool)
	pm.mu.Unlock()

	for _, pool := range pools {
		pool.Release()
	}
}

// ActiveCount returns the number of active pools.
func (pm *PoolManager) ActiveCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.pools)
}

// calculateHealth determines pool health based on utilization percentage.
// idle: <10%, active: 10-80%, saturated: >80%
func calculateHealth(pool *ants.Pool) PoolHealth {
	cap := pool.Cap()
	if cap == 0 {
		return PoolHealthIdle
	}

	utilization := float64(pool.Running()) / float64(cap) * 100

	switch {
	case utilization > 80:
		return PoolHealthSaturated
	case utilization >= 10:
		return PoolHealthActive
	default:
		return PoolHealthIdle
	}
}
