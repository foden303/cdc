package flow

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/foden/cdc/internal/core/domain"
	"github.com/google/cel-go/cel"
)

// Filter evaluates a filter expression against event data.
// It supports two modes:
//   - CEL expression evaluation against JSON data payload via Evaluate(data []byte)
//   - Simple metadata field matching via Match(ev *domain.Event) (legacy)
//
// CEL expressions receive a "data" variable which is the event's after/before payload as a map.
// Example expressions: `data.status == "active"`, `data.amount > 100`
// Empty expression always returns true (pass all).
type Filter struct {
	expression string
	// CEL compiled program (nil means pass-all)
	program cel.Program
	// Legacy simple condition (nil means pass-all)
	cond *condition
}

// operator represents a comparison operator for legacy simple expressions.
type operator int

const (
	opEq operator = iota
	opNe
)

// condition is a compiled legacy filter condition.
type condition struct {
	field string
	op    operator
	value string
}

// NewFilter parses a filter expression and returns a compiled Filter.
// An empty expression creates a pass-all filter.
// The expression is compiled as a CEL expression with a "data" variable of type map[string]dyn.
// Returns an error if the expression syntax is invalid.
func NewFilter(expression string) (*Filter, error) {
	expr := strings.TrimSpace(expression)
	if expr == "" {
		return &Filter{expression: ""}, nil
	}

	// Try to parse as a CEL expression first
	program, err := compileCEL(expr)
	if err != nil {
		// Fall back to legacy simple expression parsing for backward compatibility
		cond, legacyErr := parseExpression(expr)
		if legacyErr != nil {
			// Return the CEL error as it's the primary engine
			return nil, fmt.Errorf("filter: invalid expression %q: %w", expr, err)
		}
		return &Filter{
			expression: expr,
			cond:       cond,
		}, nil
	}

	return &Filter{
		expression: expr,
		program:    program,
	}, nil
}

// compileCEL compiles a CEL expression with a "data" variable of type map[string]dyn.
func compileCEL(expr string) (cel.Program, error) {
	env, err := cel.NewEnv(
		cel.Variable("data", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("CEL compile error: %w", issues.Err())
	}

	program, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("CEL program error: %w", err)
	}

	return program, nil
}

// Evaluate evaluates the filter expression against the event's JSON data payload.
// Returns true if the event should pass through, false if it should be skipped.
// A nil filter or empty expression always returns true (pass all).
// If data is nil or empty, and a filter expression is set, returns false.
func (f *Filter) Evaluate(data []byte) bool {
	if f == nil || (f.program == nil && f.cond == nil) {
		return true
	}

	if len(data) == 0 {
		return false
	}

	// If we have a CEL program, evaluate it
	if f.program != nil {
		return f.evaluateCEL(data)
	}

	// Legacy mode: can't evaluate simple conditions against raw data
	// Simple conditions only work with Match(ev)
	return true
}

// evaluateCEL evaluates the CEL program against JSON data.
func (f *Filter) evaluateCEL(data []byte) bool {
	// Parse JSON data into a map
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		// If data can't be parsed as JSON, skip the event
		return false
	}

	// Evaluate the CEL expression
	out, _, err := f.program.Eval(map[string]interface{}{
		"data": dataMap,
	})
	if err != nil {
		// On evaluation error, skip the event
		return false
	}

	// The result must be a boolean
	result, ok := out.Value().(bool)
	if !ok {
		return false
	}

	return result
}

// Match evaluates the filter against an event's metadata fields (legacy mode).
// Returns true if the event passes the filter (should be processed).
// A nil filter or empty expression always returns true.
func (f *Filter) Match(ev *domain.Event) bool {
	if f == nil || (f.cond == nil && f.program == nil) {
		return true
	}

	// If we have a legacy condition, use it
	if f.cond != nil {
		if ev == nil {
			return false
		}
		fieldValue := getEventField(ev, f.cond.field)
		switch f.cond.op {
		case opEq:
			return fieldValue == f.cond.value
		case opNe:
			return fieldValue != f.cond.value
		default:
			return true
		}
	}

	// If we have a CEL program but no legacy condition, try evaluating against event data
	if f.program != nil && ev != nil {
		return f.Evaluate(ev.Data)
	}

	return true
}

// Expression returns the original filter expression string.
func (f *Filter) Expression() string {
	if f == nil {
		return ""
	}
	return f.expression
}

// parseExpression parses a "field op value" expression into a condition (legacy).
func parseExpression(expr string) (*condition, error) {
	// Try != first (longer operator) to avoid matching == inside !=
	if idx := strings.Index(expr, "!="); idx >= 0 {
		field := strings.TrimSpace(expr[:idx])
		value := strings.TrimSpace(expr[idx+2:])
		if field == "" {
			return nil, fmt.Errorf("filter: empty field name in expression %q", expr)
		}
		if err := validateField(field); err != nil {
			return nil, err
		}
		return &condition{
			field: field,
			op:    opNe,
			value: stripQuotes(value),
		}, nil
	}

	if idx := strings.Index(expr, "=="); idx >= 0 {
		field := strings.TrimSpace(expr[:idx])
		value := strings.TrimSpace(expr[idx+2:])
		if field == "" {
			return nil, fmt.Errorf("filter: empty field name in expression %q", expr)
		}
		if err := validateField(field); err != nil {
			return nil, err
		}
		return &condition{
			field: field,
			op:    opEq,
			value: stripQuotes(value),
		}, nil
	}

	return nil, fmt.Errorf("filter: unsupported expression syntax %q, expected \"field == value\" or \"field != value\"", expr)
}

// supportedFields lists the event metadata fields that can be filtered on (legacy mode).
var supportedFields = map[string]bool{
	"op":          true,
	"table":       true,
	"schema":      true,
	"instance_id": true,
}

// validateField checks that the field name is a supported filter field (legacy mode).
func validateField(field string) error {
	if !supportedFields[field] {
		return fmt.Errorf("filter: unsupported field %q, supported fields: op, table, schema, instance_id", field)
	}
	return nil
}

// getEventField extracts the value of a metadata field from an event.
func getEventField(ev *domain.Event, field string) string {
	if ev == nil {
		return ""
	}
	switch field {
	case "op":
		return string(ev.Op)
	case "table":
		return ev.Table
	case "schema":
		return ev.Schema
	case "instance_id":
		return ev.InstanceID
	default:
		return ""
	}
}

// stripQuotes removes surrounding single or double quotes from a value string.
func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
