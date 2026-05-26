package flow

import (
	"encoding/json"
	"testing"
)

func TestNewFilter_EmptyExpression(t *testing.T) {
	f, err := NewFilter("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if f.program != nil {
		t.Error("expected nil CEL program for empty expression")
	}
}

func TestNewFilter_WhitespaceExpression(t *testing.T) {
	f, err := NewFilter("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.program != nil {
		t.Error("expected nil CEL program for whitespace-only expression")
	}
}

func TestNewFilter_InvalidSyntax(t *testing.T) {
	cases := []string{
		"op",
		"op > c",
		"op = c",
		"== c",
		"!= c",
		`op == "c"`,
		`table == "users"`,
	}
	for _, expr := range cases {
		_, err := NewFilter(expr)
		if err == nil {
			t.Errorf("expected error for expression %q, got nil", expr)
		}
	}
}

func TestNewFilter_UnsupportedField(t *testing.T) {
	_, err := NewFilter("unknown_field == value")
	if err == nil {
		t.Fatal("expected error for unsupported field")
	}
}

func TestFilter_Expression(t *testing.T) {
	f, _ := NewFilter(`data.op == "c"`)
	if f.Expression() != `data.op == "c"` {
		t.Errorf("Expression() = %q, want %q", f.Expression(), `data.op == "c"`)
	}

	var nilFilter *Filter
	if nilFilter.Expression() != "" {
		t.Error("nil filter Expression() should return empty string")
	}
}

func TestNewFilter_CELExpression(t *testing.T) {
	f, err := NewFilter(`data.status == "active"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if f.program == nil {
		t.Error("expected non-nil CEL program")
	}
}

func TestNewFilter_CELInvalidExpression(t *testing.T) {
	_, err := NewFilter(`data.status ++ "active"`)
	if err == nil {
		t.Fatal("expected error for invalid CEL expression")
	}
}

func TestFilter_Evaluate_PassAll(t *testing.T) {
	f, _ := NewFilter("")
	data := []byte(`{"status": "active"}`)
	if !f.Evaluate(data) {
		t.Error("empty filter should pass all events")
	}
}

func TestFilter_Evaluate_NilFilter(t *testing.T) {
	var f *Filter
	data := []byte(`{"status": "active"}`)
	if !f.Evaluate(data) {
		t.Error("nil filter should pass all events")
	}
}

func TestFilter_Evaluate_NilData(t *testing.T) {
	f, _ := NewFilter(`data.status == "active"`)
	if f.Evaluate(nil) {
		t.Error("nil data should not pass filter")
	}
}

func TestFilter_Evaluate_EmptyData(t *testing.T) {
	f, _ := NewFilter(`data.status == "active"`)
	if f.Evaluate([]byte{}) {
		t.Error("empty data should not pass filter")
	}
}

func TestFilter_Evaluate_StringEquality(t *testing.T) {
	f, _ := NewFilter(`data.status == "active"`)

	tests := []struct {
		data string
		want bool
	}{
		{`{"status": "active"}`, true},
		{`{"status": "inactive"}`, false},
		{`{"status": ""}`, false},
		{`{"other_field": "active"}`, false},
	}

	for _, tc := range tests {
		got := f.Evaluate([]byte(tc.data))
		if got != tc.want {
			t.Errorf("Evaluate(%s) = %v, want %v", tc.data, got, tc.want)
		}
	}
}

func TestFilter_Evaluate_NumericComparison(t *testing.T) {
	f, _ := NewFilter(`data.amount > 100`)

	tests := []struct {
		data string
		want bool
	}{
		{`{"amount": 150}`, true},
		{`{"amount": 100}`, false},
		{`{"amount": 50}`, false},
		{`{"amount": 101}`, true},
	}

	for _, tc := range tests {
		got := f.Evaluate([]byte(tc.data))
		if got != tc.want {
			t.Errorf("Evaluate(%s) = %v, want %v", tc.data, got, tc.want)
		}
	}
}

func TestFilter_Evaluate_BooleanExpression(t *testing.T) {
	f, _ := NewFilter(`data.active == true`)

	tests := []struct {
		data string
		want bool
	}{
		{`{"active": true}`, true},
		{`{"active": false}`, false},
	}

	for _, tc := range tests {
		got := f.Evaluate([]byte(tc.data))
		if got != tc.want {
			t.Errorf("Evaluate(%s) = %v, want %v", tc.data, got, tc.want)
		}
	}
}

func TestFilter_Evaluate_ComplexExpression(t *testing.T) {
	f, _ := NewFilter(`data.status == "active" && data.amount > 50`)

	tests := []struct {
		data string
		want bool
	}{
		{`{"status": "active", "amount": 100}`, true},
		{`{"status": "active", "amount": 30}`, false},
		{`{"status": "inactive", "amount": 100}`, false},
		{`{"status": "inactive", "amount": 30}`, false},
	}

	for _, tc := range tests {
		got := f.Evaluate([]byte(tc.data))
		if got != tc.want {
			t.Errorf("Evaluate(%s) = %v, want %v", tc.data, got, tc.want)
		}
	}
}

func TestFilter_Evaluate_InvalidJSON(t *testing.T) {
	f, _ := NewFilter(`data.status == "active"`)
	if f.Evaluate([]byte(`not json`)) {
		t.Error("invalid JSON should not pass filter")
	}
}

func TestFilter_Evaluate_HasField(t *testing.T) {
	f, _ := NewFilter(`has(data.email)`)

	tests := []struct {
		data string
		want bool
	}{
		{`{"email": "test@example.com"}`, true},
		{`{"name": "test"}`, false},
	}

	for _, tc := range tests {
		got := f.Evaluate([]byte(tc.data))
		if got != tc.want {
			t.Errorf("Evaluate(%s) = %v, want %v", tc.data, got, tc.want)
		}
	}
}

func TestFilter_Evaluate_WithEventData(t *testing.T) {
	f, _ := NewFilter(`data.status == "active"`)

	payload := map[string]interface{}{
		"id":     1,
		"status": "active",
		"name":   "test",
	}
	data, _ := json.Marshal(payload)

	if !f.Evaluate(data) {
		t.Error("expected event with active status to pass filter")
	}
}
