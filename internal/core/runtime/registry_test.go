package runtime

import (
	"testing"

	"github.com/foden/cdc/internal/core/ports"
)

func TestParseTableRef(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		schema string
		table  string
	}{
		{name: "schema table", input: "public.orders", schema: "public", table: "orders"},
		{name: "table only", input: "orders", schema: "", table: "orders"},
		{name: "empty", input: "", schema: "", table: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTableRef(tt.input)
			if got.Schema != tt.schema || got.Name != tt.table {
				t.Fatalf("ParseTableRef(%q) = %#v, want schema=%q table=%q", tt.input, got, tt.schema, tt.table)
			}
		})
	}
}

func TestNewFlowRuntimeInfoResolvesDefaults(t *testing.T) {
	info, err := NewFlowRuntimeInfo(&ports.FlowConfig{
		FlowID:      "flow-1",
		SourceID:    "source-1",
		SinkID:      "sink-1",
		SourceTable: "public.orders",
		SinkTable:   "warehouse.orders",
	})
	if err != nil {
		t.Fatalf("NewFlowRuntimeInfo failed: %v", err)
	}

	if info.SourceTable.Schema != "public" || info.SourceTable.Name != "orders" {
		t.Fatalf("unexpected source table: %#v", info.SourceTable)
	}
	if info.SinkTable.Schema != "warehouse" || info.SinkTable.Name != "orders" {
		t.Fatalf("unexpected sink table: %#v", info.SinkTable)
	}
	if info.PartitionCount != 4 {
		t.Fatalf("PartitionCount = %d, want 4", info.PartitionCount)
	}
	if info.BatchSize != 100 {
		t.Fatalf("BatchSize = %d, want 100", info.BatchSize)
	}
	if info.FlushIntervalMs != 1000 {
		t.Fatalf("FlushIntervalMs = %d, want 1000", info.FlushIntervalMs)
	}
}

func TestRegistryRegisterLookupAndUnregisterFlow(t *testing.T) {
	reg := NewRegistry()
	flow := &ports.FlowConfig{
		FlowID:      "flow-1",
		SourceID:    "source-1",
		SinkID:      "sink-1",
		SourceTable: "public.orders",
		SinkTable:   "public.orders",
		Options:     &ports.FlowOptions{PartitionCount: 8},
	}

	if err := reg.RegisterFlow(flow); err != nil {
		t.Fatalf("RegisterFlow failed: %v", err)
	}

	info, ok := reg.LookupFlow("flow-1")
	if !ok {
		t.Fatal("expected flow runtime info")
	}
	if info.PartitionCount != 8 {
		t.Fatalf("PartitionCount = %d, want 8", info.PartitionCount)
	}

	interest, ok := reg.LookupTable("source-1", "public", "orders")
	if !ok || !interest.Active {
		t.Fatalf("expected active table interest, got %#v ok=%v", interest, ok)
	}
	if interest.PartitionCount != 8 || interest.FlowCount != 1 {
		t.Fatalf("interest = %#v, want partition=8 flow_count=1", interest)
	}

	reg.UnregisterFlow("flow-1")
	if _, ok := reg.LookupFlow("flow-1"); ok {
		t.Fatal("expected flow to be removed")
	}
	interest, ok = reg.LookupTable("source-1", "public", "orders")
	if ok || interest.Active {
		t.Fatalf("expected inactive table, got %#v ok=%v", interest, ok)
	}
}

func TestRegistryRecomputesMaxPartitionOnUnregister(t *testing.T) {
	reg := NewRegistry()
	flows := []*ports.FlowConfig{
		{FlowID: "small", SourceID: "source-1", SinkID: "sink-1", SourceTable: "public.orders", SinkTable: "public.orders", Options: &ports.FlowOptions{PartitionCount: 4}},
		{FlowID: "large", SourceID: "source-1", SinkID: "sink-2", SourceTable: "public.orders", SinkTable: "public.orders", Options: &ports.FlowOptions{PartitionCount: 16}},
	}

	for _, flow := range flows {
		if err := reg.RegisterFlow(flow); err != nil {
			t.Fatalf("RegisterFlow(%s) failed: %v", flow.FlowID, err)
		}
	}

	interest, _ := reg.LookupTable("source-1", "public", "orders")
	if interest.PartitionCount != 16 || interest.FlowCount != 2 {
		t.Fatalf("interest before unregister = %#v", interest)
	}

	reg.UnregisterFlow("large")
	interest, _ = reg.LookupTable("source-1", "public", "orders")
	if interest.PartitionCount != 4 || interest.FlowCount != 1 {
		t.Fatalf("interest after unregister = %#v, want partition=4 flow_count=1", interest)
	}
}
