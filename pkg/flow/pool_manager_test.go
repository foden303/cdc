package flow

import (
	"sync"
	"testing"
)

func TestNewPoolManager(t *testing.T) {
	pm := NewPoolManager()
	if pm == nil {
		t.Fatal("expected non-nil PoolManager")
	}
	if pm.ActiveCount() != 0 {
		t.Fatalf("expected 0 active pools, got %d", pm.ActiveCount())
	}
}

func TestCreatePool(t *testing.T) {
	pm := NewPoolManager()

	pool, err := pm.CreatePool("flow-1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	if pm.ActiveCount() != 1 {
		t.Fatalf("expected 1 active pool, got %d", pm.ActiveCount())
	}

	// Create a second pool
	pool2, err := pm.CreatePool("flow-2", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool2 == nil {
		t.Fatal("expected non-nil pool")
	}
	if pm.ActiveCount() != 2 {
		t.Fatalf("expected 2 active pools, got %d", pm.ActiveCount())
	}

	// Cleanup
	pm.PauseAll()
}

func TestCreatePoolInvalidSize(t *testing.T) {
	pm := NewPoolManager()

	_, err := pm.CreatePool("flow-bad", -1)
	if err == nil {
		t.Fatal("expected error for negative pool size")
	}
}

func TestReleasePool(t *testing.T) {
	pm := NewPoolManager()

	_, err := pm.CreatePool("flow-1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pm.ReleasePool("flow-1")
	if pm.ActiveCount() != 0 {
		t.Fatalf("expected 0 active pools after release, got %d", pm.ActiveCount())
	}

	// Releasing a non-existent pool should not panic
	pm.ReleasePool("non-existent")
}

func TestGetMetrics(t *testing.T) {
	pm := NewPoolManager()

	_, err := pm.CreatePool("flow-1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer pm.PauseAll()

	metrics := pm.GetMetrics("flow-1")
	if metrics == nil {
		t.Fatal("expected non-nil metrics")
	}
	if metrics.FlowID != "flow-1" {
		t.Fatalf("expected flow_id 'flow-1', got '%s'", metrics.FlowID)
	}
	if metrics.PoolCapacity != 10 {
		t.Fatalf("expected pool capacity 10, got %d", metrics.PoolCapacity)
	}
	if metrics.Health != PoolHealthIdle {
		t.Fatalf("expected health 'idle', got '%s'", metrics.Health)
	}

	// Non-existent flow returns nil
	nilMetrics := pm.GetMetrics("non-existent")
	if nilMetrics != nil {
		t.Fatal("expected nil metrics for non-existent flow")
	}
}

func TestGetAllMetrics(t *testing.T) {
	pm := NewPoolManager()

	_, _ = pm.CreatePool("flow-1", 10)
	_, _ = pm.CreatePool("flow-2", 5)
	defer pm.PauseAll()

	allMetrics := pm.GetAllMetrics()
	if len(allMetrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(allMetrics))
	}

	flowIDs := make(map[string]bool)
	for _, m := range allMetrics {
		flowIDs[m.FlowID] = true
	}
	if !flowIDs["flow-1"] || !flowIDs["flow-2"] {
		t.Fatal("expected metrics for both flow-1 and flow-2")
	}
}

func TestGetHealth(t *testing.T) {
	pm := NewPoolManager()

	_, err := pm.CreatePool("flow-1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer pm.PauseAll()

	// No tasks running, should be idle
	health := pm.GetHealth("flow-1")
	if health != PoolHealthIdle {
		t.Fatalf("expected health 'idle', got '%s'", health)
	}

	// Non-existent flow returns idle
	health = pm.GetHealth("non-existent")
	if health != PoolHealthIdle {
		t.Fatalf("expected health 'idle' for non-existent flow, got '%s'", health)
	}
}

func TestPauseAll(t *testing.T) {
	pm := NewPoolManager()

	_, _ = pm.CreatePool("flow-1", 10)
	_, _ = pm.CreatePool("flow-2", 5)
	_, _ = pm.CreatePool("flow-3", 8)

	if pm.ActiveCount() != 3 {
		t.Fatalf("expected 3 active pools, got %d", pm.ActiveCount())
	}

	pm.PauseAll()

	if pm.ActiveCount() != 0 {
		t.Fatalf("expected 0 active pools after PauseAll, got %d", pm.ActiveCount())
	}
}

func TestActiveCount(t *testing.T) {
	pm := NewPoolManager()

	if pm.ActiveCount() != 0 {
		t.Fatalf("expected 0, got %d", pm.ActiveCount())
	}

	_, _ = pm.CreatePool("flow-1", 10)
	if pm.ActiveCount() != 1 {
		t.Fatalf("expected 1, got %d", pm.ActiveCount())
	}

	_, _ = pm.CreatePool("flow-2", 5)
	if pm.ActiveCount() != 2 {
		t.Fatalf("expected 2, got %d", pm.ActiveCount())
	}

	pm.ReleasePool("flow-1")
	if pm.ActiveCount() != 1 {
		t.Fatalf("expected 1, got %d", pm.ActiveCount())
	}

	pm.ReleasePool("flow-2")
	if pm.ActiveCount() != 0 {
		t.Fatalf("expected 0, got %d", pm.ActiveCount())
	}
}

func TestConcurrentAccess(t *testing.T) {
	pm := NewPoolManager()
	defer pm.PauseAll()

	var wg sync.WaitGroup
	// Concurrently create pools
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			flowID := "flow-" + string(rune('a'+id))
			_, _ = pm.CreatePool(flowID, 5)
		}(i)
	}
	wg.Wait()

	// Concurrently read metrics
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pm.GetAllMetrics()
			_ = pm.ActiveCount()
		}()
	}
	wg.Wait()
}

func TestCalculateHealth(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
		expected PoolHealth
	}{
		{"idle pool", 10, PoolHealthIdle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPoolManager()
			_, err := pm.CreatePool("test-flow", tt.poolSize)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer pm.PauseAll()

			health := pm.GetHealth("test-flow")
			if health != tt.expected {
				t.Fatalf("expected health '%s', got '%s'", tt.expected, health)
			}
		})
	}
}
