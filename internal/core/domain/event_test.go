package domain

import (
	"testing"
)

func TestEvent_Reset(t *testing.T) {
	e := &Event{
		Topic:      "topic",
		Subject:    "subject",
		InstanceID: "inst-1",
		Schema:     "public",
		Table:      "users",
		Op:         "c",
		Offset:     "0/1234",
		LSN:        12345,
		Data:       []byte(`{"key":"value"}`),
		MessageID:  "msg-1",
		Partition:  3,
	}

	e.Reset()

	if e.Topic != "" {
		t.Errorf("Topic not reset: got %q", e.Topic)
	}
	if e.Subject != "" {
		t.Errorf("Subject not reset: got %q", e.Subject)
	}
	if e.InstanceID != "" {
		t.Errorf("InstanceID not reset: got %q", e.InstanceID)
	}
	if e.Schema != "" {
		t.Errorf("Schema not reset: got %q", e.Schema)
	}
	if e.Table != "" {
		t.Errorf("Table not reset: got %q", e.Table)
	}
	if e.Op != "" {
		t.Errorf("Op not reset: got %q", e.Op)
	}
	if e.Offset != "" {
		t.Errorf("Offset not reset: got %q", e.Offset)
	}
	if e.LSN != 0 {
		t.Errorf("LSN not reset: got %d", e.LSN)
	}
	if e.Data != nil {
		t.Errorf("Data not reset: got %v", e.Data)
	}
	if e.MessageID != "" {
		t.Errorf("MessageID not reset: got %q", e.MessageID)
	}
	if e.Partition != 0 {
		t.Errorf("Partition not reset: got %d", e.Partition)
	}
}

func TestEvent_DeepClone(t *testing.T) {
	original := &Event{
		Topic:      "topic",
		Subject:    "subject",
		InstanceID: "inst-1",
		Schema:     "public",
		Table:      "users",
		Op:         "c",
		Offset:     "0/1234",
		LSN:        12345,
		Data:       []byte(`{"key":"value"}`),
		MessageID:  "msg-1",
		Partition:  3,
	}

	clone := original.DeepClone()

	// Verify all scalar fields are copied
	if clone.Topic != original.Topic {
		t.Errorf("Topic mismatch: got %q, want %q", clone.Topic, original.Topic)
	}
	if clone.Subject != original.Subject {
		t.Errorf("Subject mismatch: got %q, want %q", clone.Subject, original.Subject)
	}
	if clone.InstanceID != original.InstanceID {
		t.Errorf("InstanceID mismatch: got %q, want %q", clone.InstanceID, original.InstanceID)
	}
	if clone.Schema != original.Schema {
		t.Errorf("Schema mismatch: got %q, want %q", clone.Schema, original.Schema)
	}
	if clone.Table != original.Table {
		t.Errorf("Table mismatch: got %q, want %q", clone.Table, original.Table)
	}
	if clone.Op != original.Op {
		t.Errorf("Op mismatch: got %q, want %q", clone.Op, original.Op)
	}
	if clone.Offset != original.Offset {
		t.Errorf("Offset mismatch: got %q, want %q", clone.Offset, original.Offset)
	}
	if clone.LSN != original.LSN {
		t.Errorf("LSN mismatch: got %d, want %d", clone.LSN, original.LSN)
	}
	if clone.MessageID != original.MessageID {
		t.Errorf("MessageID mismatch: got %q, want %q", clone.MessageID, original.MessageID)
	}
	if clone.Partition != original.Partition {
		t.Errorf("Partition mismatch: got %d, want %d", clone.Partition, original.Partition)
	}

	// Verify Data is deeply copied (no shared backing array)
	if string(clone.Data) != string(original.Data) {
		t.Errorf("Data content mismatch: got %q, want %q", clone.Data, original.Data)
	}

	// Modify original Data and verify clone is unaffected
	original.Data[0] = 'X'
	if clone.Data[0] == 'X' {
		t.Error("DeepClone shares backing array with original: modifying original affected clone")
	}
}

func TestEvent_DeepClone_NilData(t *testing.T) {
	original := &Event{
		Topic:      "topic",
		InstanceID: "inst-1",
		Schema:     "public",
		Table:      "users",
		Op:         "c",
		Data:       nil,
	}

	clone := original.DeepClone()

	if clone.Data != nil {
		t.Errorf("Expected nil Data in clone, got %v", clone.Data)
	}
}
