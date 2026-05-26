package runtime

import (
	"fmt"
	"strings"

	"github.com/foden/cdc/internal/core/ports"
)

const (
	defaultPartitionCount  = int32(4)
	defaultBatchSize       = int32(100)
	defaultFlushIntervalMs = int32(1000)
)

type TableRef struct {
	Schema string
	Name   string
}

type FlowRuntimeInfo struct {
	FlowID   string
	SourceID string
	SinkID   string

	SourceTable TableRef
	SinkTable   TableRef

	FilterExpression string
	ColumnMappings   []ports.ColumnMapping

	BatchSize       int32
	FlushIntervalMs int32
	PartitionCount  int32
}

type tableKey struct {
	sourceID string
	schema   string
	table    string
}

func ParseTableRef(value string) TableRef {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) == 2 {
		return TableRef{Schema: parts[0], Name: parts[1]}
	}
	return TableRef{Name: value}
}

func newTableKey(sourceID string, table TableRef) tableKey {
	return tableKey{sourceID: sourceID, schema: table.Schema, table: table.Name}
}

func NewFlowRuntimeInfo(flow *ports.FlowConfig) (*FlowRuntimeInfo, error) {
	if flow == nil {
		return nil, fmt.Errorf("flow config is required")
	}
	if flow.FlowID == "" {
		return nil, fmt.Errorf("flow_id is required")
	}
	if flow.SourceID == "" {
		return nil, fmt.Errorf("source_id is required")
	}
	if flow.SinkID == "" {
		return nil, fmt.Errorf("sink_id is required")
	}
	if flow.SourceTable == "" {
		return nil, fmt.Errorf("source_table is required")
	}
	if flow.SinkTable == "" {
		return nil, fmt.Errorf("sink_table is required")
	}

	info := &FlowRuntimeInfo{
		FlowID:           flow.FlowID,
		SourceID:         flow.SourceID,
		SinkID:           flow.SinkID,
		SourceTable:      ParseTableRef(flow.SourceTable),
		SinkTable:        ParseTableRef(flow.SinkTable),
		ColumnMappings:   flow.ColumnMappings,
		PartitionCount:   defaultPartitionCount,
		BatchSize:        defaultBatchSize,
		FlushIntervalMs:  defaultFlushIntervalMs,
		FilterExpression: "",
	}

	if flow.Options != nil {
		info.FilterExpression = flow.Options.FilterExpression
		if flow.Options.PartitionCount > 0 {
			info.PartitionCount = int32(flow.Options.PartitionCount)
		}
		if flow.Options.BatchSize > 0 {
			info.BatchSize = flow.Options.BatchSize
		}
		if flow.Options.FlushIntervalMs > 0 {
			info.FlushIntervalMs = flow.Options.FlushIntervalMs
		}
	}

	return info, nil
}
