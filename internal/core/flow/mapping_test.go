package flow

import (
	"testing"

	"github.com/bytedance/sonic"
	"github.com/foden/cdc/internal/core/ports"
)

func TestApplyColumnMappings_RenameEnabled(t *testing.T) {
	data := []byte(`{"id":1,"name":"Alice","age":30}`)
	mappings := []ports.ColumnMapping{
		{SourceColumn: "id", SinkColumn: "user_id", Enabled: true},
		{SourceColumn: "name", SinkColumn: "full_name", Enabled: true},
	}

	result, err := ApplyColumnMappings(data, mappings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]interface{}
	if err := sonic.Unmarshal(result, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// "id" should be renamed to "user_id"
	if _, ok := out["user_id"]; !ok {
		t.Error("expected 'user_id' key in output")
	}
	// "name" should be renamed to "full_name"
	if _, ok := out["full_name"]; !ok {
		t.Error("expected 'full_name' key in output")
	}
	// "age" is not in any mapping, should pass through unchanged
	if _, ok := out["age"]; !ok {
		t.Error("expected 'age' key to pass through unchanged")
	}
	// Original keys should not exist
	if _, ok := out["id"]; ok {
		t.Error("'id' should have been renamed to 'user_id'")
	}
	if _, ok := out["name"]; ok {
		t.Error("'name' should have been renamed to 'full_name'")
	}
}

func TestApplyColumnMappings_ExcludeDisabled(t *testing.T) {
	data := []byte(`{"id":1,"secret":"hidden","name":"Bob"}`)
	mappings := []ports.ColumnMapping{
		{SourceColumn: "id", SinkColumn: "id", Enabled: true},
		{SourceColumn: "secret", SinkColumn: "secret", Enabled: false},
	}

	result, err := ApplyColumnMappings(data, mappings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]interface{}
	if err := sonic.Unmarshal(result, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// "secret" should be excluded (disabled)
	if _, ok := out["secret"]; ok {
		t.Error("'secret' should be excluded (disabled mapping)")
	}
	// "id" should be present (enabled)
	if _, ok := out["id"]; !ok {
		t.Error("expected 'id' key in output")
	}
	// "name" is not in any mapping, should pass through
	if _, ok := out["name"]; !ok {
		t.Error("expected 'name' to pass through unchanged")
	}
}

func TestApplyColumnMappings_EmptyMappings(t *testing.T) {
	data := []byte(`{"id":1,"name":"Charlie"}`)

	result, err := ApplyColumnMappings(data, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]interface{}
	if err := sonic.Unmarshal(result, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// All columns should pass through unchanged
	if _, ok := out["id"]; !ok {
		t.Error("expected 'id' to pass through")
	}
	if _, ok := out["name"]; !ok {
		t.Error("expected 'name' to pass through")
	}
}

func TestApplyColumnMappings_InvalidJSON(t *testing.T) {
	data := []byte(`not valid json`)
	mappings := []ports.ColumnMapping{
		{SourceColumn: "id", SinkColumn: "id", Enabled: true},
	}

	_, err := ApplyColumnMappings(data, mappings)
	if err == nil {
		t.Error("expected error for invalid JSON input")
	}
}

func TestAutoGenerateMappings_CaseInsensitiveMatch(t *testing.T) {
	sourceColumns := []ports.ColumnInfo{
		{Name: "ID", Type: "integer"},
		{Name: "UserName", Type: "text"},
		{Name: "email", Type: "text"},
		{Name: "extra_col", Type: "text"},
	}
	sinkColumns := []ports.ColumnInfo{
		{Name: "id", Type: "bigint"},
		{Name: "username", Type: "varchar"},
		{Name: "Email", Type: "varchar"},
		{Name: "other_col", Type: "text"},
	}

	mappings := AutoGenerateMappings(sourceColumns, sinkColumns)

	if len(mappings) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(mappings))
	}

	// Verify each mapping
	expected := map[string]struct {
		sinkCol    string
		sourceType string
		sinkType   string
	}{
		"ID":       {sinkCol: "id", sourceType: "integer", sinkType: "bigint"},
		"UserName": {sinkCol: "username", sourceType: "text", sinkType: "varchar"},
		"email":    {sinkCol: "Email", sourceType: "text", sinkType: "varchar"},
	}

	for _, m := range mappings {
		exp, ok := expected[m.SourceColumn]
		if !ok {
			t.Errorf("unexpected mapping for source column: %s", m.SourceColumn)
			continue
		}
		if m.SinkColumn != exp.sinkCol {
			t.Errorf("for %s: expected sink_column=%s, got %s", m.SourceColumn, exp.sinkCol, m.SinkColumn)
		}
		if m.SourceType != exp.sourceType {
			t.Errorf("for %s: expected source_type=%s, got %s", m.SourceColumn, exp.sourceType, m.SourceType)
		}
		if m.SinkType != exp.sinkType {
			t.Errorf("for %s: expected sink_type=%s, got %s", m.SourceColumn, exp.sinkType, m.SinkType)
		}
		if !m.Enabled {
			t.Errorf("for %s: expected enabled=true", m.SourceColumn)
		}
	}
}

func TestAutoGenerateMappings_NoMatches(t *testing.T) {
	sourceColumns := []ports.ColumnInfo{
		{Name: "foo", Type: "text"},
	}
	sinkColumns := []ports.ColumnInfo{
		{Name: "bar", Type: "text"},
	}

	mappings := AutoGenerateMappings(sourceColumns, sinkColumns)

	if len(mappings) != 0 {
		t.Errorf("expected 0 mappings for non-matching columns, got %d", len(mappings))
	}
}

func TestAutoGenerateMappings_EmptyInputs(t *testing.T) {
	mappings := AutoGenerateMappings(nil, nil)
	if mappings != nil {
		t.Errorf("expected nil for empty inputs, got %v", mappings)
	}

	mappings = AutoGenerateMappings([]ports.ColumnInfo{}, []ports.ColumnInfo{})
	if mappings != nil {
		t.Errorf("expected nil for empty slices, got %v", mappings)
	}
}
