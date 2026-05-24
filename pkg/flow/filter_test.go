package flow

import (
	"encoding/json"
	"testing"

	"github.com/foden/cdc/pkg/models"
)

func TestNewFilter_EmptyExpression(t *testing.T) {
	f, err := NewFilter("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if f.cond != nil {
		t.Error("expected nil condition for empty expression")
	}
}

func TestNewFilter_WhitespaceExpression(t *testing.T) {
	f, err := NewFilter("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.cond != nil {
		t.Error("expected nil condition for whitespace-only expression")
	}
}

func TestNewFilter_InvalidSyntax(t *testing.T) {
	cases := []string{
		"op",
		"op > c",
		"op = c",
		"== c",
		"!= c",
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

func TestNewFilter_ValidExpressions(t *testing.T) {
	cases := []struct {
		expr  string
		field string
		op    operator
		value string
	}{
		{`op == "c"`, "op", opEq, "c"},
		{`op == c`, "op", opEq, "c"},
		{`table == "users"`, "table", opEq, "users"},
		{`schema == public`, "schema", opEq, "public"},
		{`instance_id != "abc-123"`, "instance_id", opNe, "abc-123"},
		{`op != d`, "op", opNe, "d"},
		{`table == 'orders'`, "table", opEq, "orders"},
	}
	for _, tc := range cases {
		f, err := NewFilter(tc.expr)
		if err != nil {
			t.Errorf("NewFilter(%q) error: %v", tc.expr, err)
			continue
		}
		if f.cond == nil {
			t.Errorf("NewFilter(%q) cond is nil", tc.expr)
			continue
		}
		if f.cond.field != tc.field {
			t.Errorf("NewFilter(%q) field = %q, want %q", tc.expr, f.cond.field, tc.field)
		}
		if f.cond.op != tc.op {
			t.Errorf("NewFilter(%q) op = %v, want %v", tc.expr, f.cond.op, tc.op)
		}
		if f.cond.value != tc.value {
			t.Errorf("NewFilter(%q) value = %q, want %q", tc.expr, f.cond.value, tc.value)
		}
	}
}

func TestFilter_Match_PassAll(t *testing.T) {
	f, _ := NewFilter("")
	ev := &models.Event{Op: "c", Table: "users", Schema: "public", InstanceID: "src-1"}

	if !f.Match(ev) {
		t.Error("empty filter should pass all events")
	}
}

func TestFilter_Match_NilFilter(t *testing.T) {
	var f *Filter
	ev := &models.Event{Op: "c", Table: "users", Schema: "public"}

	if !f.Match(ev) {
		t.Error("nil filter should pass all events")
	}
}

func TestFilter_Match_NilEvent(t *testing.T) {
	f, _ := NewFilter("op == c")
	if f.Match(nil) {
		t.Error("nil event should not match any condition")
	}
}

func TestFilter_Match_EqualOperator(t *testing.T) {
	f, _ := NewFilter(`op == "c"`)

	tests := []struct {
		event *models.Event
		want  bool
	}{
		{&models.Event{Op: "c"}, true},
		{&models.Event{Op: "u"}, false},
		{&models.Event{Op: "d"}, false},
		{&models.Event{Op: ""}, false},
	}

	for _, tc := range tests {
		got := f.Match(tc.event)
		if got != tc.want {
			t.Errorf("Match(op=%q) = %v, want %v", tc.event.Op, got, tc.want)
		}
	}
}

func TestFilter_Match_NotEqualOperator(t *testing.T) {
	f, _ := NewFilter(`op != "d"`)

	tests := []struct {
		event *models.Event
		want  bool
	}{
		{&models.Event{Op: "c"}, true},
		{&models.Event{Op: "u"}, true},
		{&models.Event{Op: "d"}, false},
	}

	for _, tc := range tests {
		got := f.Match(tc.event)
		if got != tc.want {
			t.Errorf("Match(op=%q) = %v, want %v", tc.event.Op, got, tc.want)
		}
	}
}

func TestFilter_Match_TableField(t *testing.T) {
	f, _ := NewFilter(`table == "users"`)

	if !f.Match(&models.Event{Table: "users"}) {
		t.Error("expected match for table=users")
	}
	if f.Match(&models.Event{Table: "orders"}) {
		t.Error("expected no match for table=orders")
	}
}

func TestFilter_Match_SchemaField(t *testing.T) {
	f, _ := NewFilter(`schema == "public"`)

	if !f.Match(&models.Event{Schema: "public"}) {
		t.Error("expected match for schema=public")
	}
	if f.Match(&models.Event{Schema: "private"}) {
		t.Error("expected no match for schema=private")
	}
}

func TestFilter_Match_InstanceIDField(t *testing.T) {
	f, _ := NewFilter(`instance_id == "src-abc"`)

	if !f.Match(&models.Event{InstanceID: "src-abc"}) {
		t.Error("expected match for instance_id=src-abc")
	}
	if f.Match(&models.Event{InstanceID: "src-xyz"}) {
		t.Error("expected no match for instance_id=src-xyz")
	}
}

func TestFilter_Expression(t *testing.T) {
	f, _ := NewFilter(`op == "c"`)
	if f.Expression() != `op == "c"` {
		t.Errorf("Expression() = %q, want %q", f.Expression(), `op == "c"`)
	}

	var nilFilter *Filter
	if nilFilter.Expression() != "" {
		t.Error("nil filter Expression() should return empty string")
	}
}

func TestStripQuotes(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`hello`, "hello"},
		{`"`, `"`},
		{`""`, ""},
		{`''`, ""},
		{`"mixed'`, `"mixed'`},
	}
	for _, tc := range cases {
		got := stripQuotes(tc.input)
		if got != tc.want {
			t.Errorf("stripQuotes(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// --- CEL Expression Evaluate Tests ---

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
	// Simulate how the worker would use Evaluate with event.Data
	f, _ := NewFilter(`data.status == "active"`)

	payload := map[string]interface{}{
		"id":     1,
		"status": "active",
		"name":   "test",
	}
	data, _ := json.Marshal(payload)

	ev := &models.Event{
		Op:    "c",
		Table: "users",
		Data:  data,
	}

	if !f.Evaluate(ev.Data) {
		t.Error("expected event with active status to pass filter")
	}
}
