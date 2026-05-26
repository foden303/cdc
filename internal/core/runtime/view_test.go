package runtime

import (
	"testing"

	"github.com/foden/cdc/internal/core/ports"
)

func TestViewFlowStatsCombinesMetricsAndRuntimeInfo(t *testing.T) {
	reg := NewRegistry()
	metrics := NewMetrics()
	view := NewView(reg, metrics, nil)

	flow := &ports.FlowConfig{
		FlowID:      "flow-1",
		SourceID:    "source-1",
		SinkID:      "sink-1",
		SourceTable: "public.orders",
		SinkTable:   "public.orders",
	}
	if err := reg.RegisterFlow(flow); err != nil {
		t.Fatalf("RegisterFlow failed: %v", err)
	}
	metrics.RecordSinkWrite("flow-1", "source-1", "sink-1", 7, 10, 1000)

	stats, ok := view.FlowStats("flow-1")
	if !ok {
		t.Fatal("expected flow stats")
	}
	if stats.TotalEventsProcessed != 7 {
		t.Fatalf("TotalEventsProcessed = %d, want 7", stats.TotalEventsProcessed)
	}
}

func TestViewDashboardAggregatesFlowStats(t *testing.T) {
	reg := NewRegistry()
	metrics := NewMetrics()
	view := NewView(reg, metrics, nil)

	_ = reg.RegisterFlow(&ports.FlowConfig{FlowID: "flow-1", SourceID: "source-1", SinkID: "sink-1", SourceTable: "public.orders", SinkTable: "public.orders"})
	metrics.RecordSinkWrite("flow-1", "source-1", "sink-1", 4, 12, 1000)

	dashboard := view.Dashboard()
	if dashboard.TotalSyncedEvents != 4 {
		t.Fatalf("TotalSyncedEvents = %d, want 4", dashboard.TotalSyncedEvents)
	}
}
