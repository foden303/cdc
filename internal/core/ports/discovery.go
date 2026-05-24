package ports

import "context"

// TableInfo represents a discovered table with its column definitions.
type TableInfo struct {
	Schema  string       `json:"schema"`
	Name    string       `json:"name"`
	Columns []ColumnInfo `json:"columns"`
}

// ColumnInfo represents a single column within a table.
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsNullable   bool   `json:"is_nullable"`
}

// Discovery provides table schema introspection for sources and sinks.
type Discovery interface {
	TestSourceConnection(ctx context.Context, cfg *SourceConfig) (latencyMs int64, err error)
	TestSinkConnection(ctx context.Context, cfg *SinkConfig) (latencyMs int64, err error)
	DiscoverSourceTables(ctx context.Context, cfg *SourceConfig) ([]TableInfo, error)
	DiscoverSinkTables(ctx context.Context, cfg *SinkConfig) ([]TableInfo, error)
}
