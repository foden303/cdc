package flow

import (
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/foden/cdc/pkg/interfaces"
	"pgregory.net/rapid"
)

// TestProperty_MappingEnabledRename verifies that for any enabled mapping,
// the output has the sink_column key with the same value as the source_column had.
// **Validates: Requirements 9.2**
func TestProperty_MappingEnabledRename(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		srcCol := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "src")
		sinkCol := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "sink")
		value := rapid.String().Draw(t, "value")

		data, err := sonic.Marshal(map[string]interface{}{srcCol: value})
		if err != nil {
			t.Fatal(err)
		}
		mappings := []interfaces.ColumnMapping{{SourceColumn: srcCol, SinkColumn: sinkCol, Enabled: true}}

		result, err := ApplyColumnMappings(data, mappings)
		if err != nil {
			t.Fatal(err)
		}

		var out map[string]interface{}
		if err := sonic.Unmarshal(result, &out); err != nil {
			t.Fatal(err)
		}

		if _, ok := out[sinkCol]; !ok {
			t.Fatalf("expected sink column %q in output, got keys: %v", sinkCol, out)
		}
		if srcCol != sinkCol {
			if _, ok := out[srcCol]; ok {
				t.Fatalf("original column %q should be renamed to %q but still present", srcCol, sinkCol)
			}
		}
	})
}

// TestProperty_MappingDisabledExclude verifies that for any disabled mapping,
// the output does NOT contain that source_column.
// **Validates: Requirements 9.3**
func TestProperty_MappingDisabledExclude(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		srcCol := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "src")
		value := rapid.String().Draw(t, "value")

		data, err := sonic.Marshal(map[string]interface{}{srcCol: value})
		if err != nil {
			t.Fatal(err)
		}
		mappings := []interfaces.ColumnMapping{{SourceColumn: srcCol, SinkColumn: srcCol, Enabled: false}}

		result, err := ApplyColumnMappings(data, mappings)
		if err != nil {
			t.Fatal(err)
		}

		var out map[string]interface{}
		if err := sonic.Unmarshal(result, &out); err != nil {
			t.Fatal(err)
		}

		if _, ok := out[srcCol]; ok {
			t.Fatalf("disabled column %q should NOT be in output, got: %v", srcCol, out)
		}
	})
}

// TestProperty_MappingUnmappedPassthrough verifies that columns not in any mapping
// appear unchanged in the output.
// **Validates: Requirements 9.4**
func TestProperty_MappingUnmappedPassthrough(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate two distinct column names
		mappedCol := rapid.StringMatching(`[a-z]{1,5}`).Draw(t, "mapped")
		unmappedCol := rapid.StringMatching(`[m-z]{1,5}`).Draw(t, "unmapped")
		// Ensure they are different
		if mappedCol == unmappedCol {
			unmappedCol = unmappedCol + "x"
		}

		mappedValue := rapid.String().Draw(t, "mappedValue")
		unmappedValue := rapid.String().Draw(t, "unmappedValue")

		data, err := sonic.Marshal(map[string]interface{}{
			mappedCol:   mappedValue,
			unmappedCol: unmappedValue,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Only map the first column, leave the second unmapped
		mappings := []interfaces.ColumnMapping{{SourceColumn: mappedCol, SinkColumn: "renamed", Enabled: true}}

		result, err := ApplyColumnMappings(data, mappings)
		if err != nil {
			t.Fatal(err)
		}

		var out map[string]interface{}
		if err := sonic.Unmarshal(result, &out); err != nil {
			t.Fatal(err)
		}

		// Unmapped column should pass through unchanged
		if val, ok := out[unmappedCol]; !ok {
			t.Fatalf("unmapped column %q should be in output, got: %v", unmappedCol, out)
		} else if val != unmappedValue {
			t.Fatalf("unmapped column %q value changed: expected %q, got %q", unmappedCol, unmappedValue, val)
		}
	})
}

// TestProperty_AutoGenerateCaseInsensitive verifies that for any pair of columns
// with the same name (different case), auto-generate produces a mapping.
// **Validates: Requirements 4.4**
func TestProperty_AutoGenerateCaseInsensitive(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		colName := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "colName")

		// Create source column in lowercase, sink column in uppercase
		sourceColumns := []interfaces.ColumnInfo{{Name: colName, Type: "text"}}
		sinkColumns := []interfaces.ColumnInfo{{Name: strings.ToUpper(colName), Type: "text"}}

		mappings := AutoGenerateMappings(sourceColumns, sinkColumns)

		if len(mappings) == 0 {
			t.Fatalf("expected auto-generated mapping for %q (source) and %q (sink)", colName, strings.ToUpper(colName))
		}

		found := false
		for _, m := range mappings {
			if strings.EqualFold(m.SourceColumn, colName) && strings.EqualFold(m.SinkColumn, strings.ToUpper(colName)) {
				found = true
				if !m.Enabled {
					t.Fatalf("auto-generated mapping should be enabled")
				}
			}
		}
		if !found {
			t.Fatalf("no mapping found matching source=%q sink=%q in %+v", colName, strings.ToUpper(colName), mappings)
		}
	})
}
