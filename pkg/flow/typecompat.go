package flow

// SinkType identifies the sink connector type for compatibility lookups.
type SinkType string

const (
	SinkTypePostgres   SinkType = "postgres"
	SinkTypeClickhouse SinkType = "clickhouse"
)

// TypeCompatibilityMatrix defines which source types can map to which sink types.
// Key: sink connector type -> map[source_type] -> list of compatible sink_types
var TypeCompatibilityMatrix = map[SinkType]map[string][]string{
	SinkTypePostgres: {
		"integer":                     {"integer", "bigint", "numeric", "text"},
		"bigint":                      {"bigint", "numeric", "text"},
		"smallint":                    {"smallint", "integer", "bigint", "numeric", "text"},
		"numeric":                     {"numeric", "text"},
		"real":                        {"real", "double precision", "numeric", "text"},
		"double precision":            {"double precision", "numeric", "text"},
		"boolean":                     {"boolean", "integer", "text"},
		"text":                        {"text", "character varying"},
		"character varying":           {"text", "character varying"},
		"uuid":                        {"uuid", "text", "character varying"},
		"timestamp without time zone": {"timestamp without time zone", "timestamp with time zone", "text"},
		"timestamp with time zone":    {"timestamp with time zone", "text"},
		"date":                        {"date", "timestamp without time zone", "text"},
		"jsonb":                       {"jsonb", "json", "text"},
		"json":                        {"json", "jsonb", "text"},
		"bytea":                       {"bytea", "text"},
	},
	SinkTypeClickhouse: {
		"integer":                     {"Int32", "Int64", "UInt32", "Float64", "String"},
		"bigint":                      {"Int64", "UInt64", "Float64", "String"},
		"smallint":                    {"Int16", "Int32", "Int64", "String"},
		"numeric":                     {"Decimal", "Float64", "String"},
		"real":                        {"Float32", "Float64", "String"},
		"double precision":            {"Float64", "String"},
		"boolean":                     {"Bool", "UInt8", "String"},
		"text":                        {"String", "FixedString"},
		"character varying":           {"String", "FixedString"},
		"uuid":                        {"UUID", "String"},
		"timestamp without time zone": {"DateTime", "DateTime64", "String"},
		"timestamp with time zone":    {"DateTime64", "String"},
		"date":                        {"Date", "Date32", "DateTime", "String"},
		"jsonb":                       {"String"},
		"json":                        {"String"},
	},
}

// PassThroughSinks are sink types where all type mappings are compatible.
var PassThroughSinks = map[SinkType]bool{}

// IsTypeCompatible checks if a source type can be mapped to a sink type
// for the given sink connector type.
func IsTypeCompatible(sinkConnectorType SinkType, sourceType, sinkType string) bool {
	// Pass-through sinks accept all types
	if PassThroughSinks[sinkConnectorType] {
		return true
	}

	matrix, ok := TypeCompatibilityMatrix[sinkConnectorType]
	if !ok {
		return false
	}

	compatibleTypes, ok := matrix[sourceType]
	if !ok {
		return false
	}

	for _, ct := range compatibleTypes {
		if ct == sinkType {
			return true
		}
	}
	return false
}

// IncompatibleMapping describes a column mapping that failed type compatibility.
type IncompatibleMapping struct {
	SourceColumn string
	SinkColumn   string
	SourceType   string
	SinkType     string
}

// ValidateColumnMappings checks all enabled mappings for type compatibility.
// Returns nil if all are compatible, or a list of incompatible mappings.
func ValidateColumnMappings(mappings []ColumnMapping, sinkConnectorType SinkType) []IncompatibleMapping {
	var incompatible []IncompatibleMapping
	for _, m := range mappings {
		if !m.Enabled {
			continue
		}
		if !IsTypeCompatible(sinkConnectorType, m.SourceType, m.SinkType) {
			incompatible = append(incompatible, IncompatibleMapping{
				SourceColumn: m.SourceColumn,
				SinkColumn:   m.SinkColumn,
				SourceType:   m.SourceType,
				SinkType:     m.SinkType,
			})
		}
	}
	return incompatible
}
