package discovery

import (
	"context"
	"testing"

	"github.com/foden/cdc/internal/core/ports"
)

func TestDiscoverSourceTablesUnsupportedTypeReturnsError(t *testing.T) {
	svc := NewService()

	tables, err := svc.DiscoverSourceTables(context.Background(), &ports.SourceConfig{Type: "oracle"})
	if err == nil {
		t.Fatal("expected unsupported source type error")
	}
	if tables != nil {
		t.Fatalf("expected nil tables, got %#v", tables)
	}
}

func TestDiscoverSinkTablesUnsupportedTypeReturnsError(t *testing.T) {
	svc := NewService()

	tables, err := svc.DiscoverSinkTables(context.Background(), &ports.SinkConfig{Type: "redis"})
	if err == nil {
		t.Fatal("expected unsupported sink type error")
	}
	if tables != nil {
		t.Fatalf("expected nil tables, got %#v", tables)
	}
}
