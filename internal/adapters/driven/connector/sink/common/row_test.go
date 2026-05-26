package common

import (
	"testing"

	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
)

func TestRowMapUsesAfterForCreateUpdateSnapshot(t *testing.T) {
	for _, op := range []constant.Op{constant.OpCreate, constant.OpUpdate, constant.OpSnapshot} {
		event := &domain.Event{
			Op:   op,
			Data: []byte(`{"before":{"id":1,"name":"old"},"after":{"id":2,"name":"new"}}`),
		}
		row, ok, err := RowMap(event)
		if err != nil {
			t.Fatalf("RowMap(%s) error: %v", op, err)
		}
		if !ok {
			t.Fatalf("RowMap(%s) ok=false", op)
		}
		if row["id"] != float64(2) || row["name"] != "new" {
			t.Fatalf("RowMap(%s) = %+v", op, row)
		}
	}
}

func TestRowMapUsesBeforeForDelete(t *testing.T) {
	event := &domain.Event{
		Op:   constant.OpDelete,
		Data: []byte(`{"before":{"id":7,"name":"old"},"after":{"id":8,"name":"new"}}`),
	}
	row, ok, err := RowMap(event)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("RowMap ok=false")
	}
	if row["id"] != float64(7) || row["name"] != "old" {
		t.Fatalf("row = %+v", row)
	}
}

func TestExtractKeyValueUsesPreferredKeys(t *testing.T) {
	key, value, ok := ExtractKeyValue(map[string]interface{}{
		"uuid": "u-1",
		"id":   float64(10),
	})
	if !ok {
		t.Fatal("ExtractKeyValue ok=false")
	}
	if key != "id" || value != float64(10) {
		t.Fatalf("key/value = %s/%v", key, value)
	}
}
