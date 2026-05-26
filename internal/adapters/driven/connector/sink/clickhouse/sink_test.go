package clickhouse

import (
	"strings"
	"testing"
	"time"

	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
)

func TestClickHouseColumnsAppendCDCMetadata(t *testing.T) {
	columns := clickhouseColumns(map[string]interface{}{"name": "Ada", "id": 1})

	wantSuffix := []string{"_cdc_op", "_cdc_ts", "_cdc_deleted", "_cdc_lsn"}
	if len(columns) != 6 {
		t.Fatalf("columns = %+v", columns)
	}
	for i, want := range wantSuffix {
		got := columns[len(columns)-len(wantSuffix)+i]
		if got != want {
			t.Fatalf("metadata column %d = %q, want %q; columns=%+v", i, got, want, columns)
		}
	}
}

func TestClickHouseAppendArgsIncludeCDCMetadata(t *testing.T) {
	event := &domain.Event{
		Op:          constant.OpUpdate,
		LSN:         12345,
		TimestampMS: 1710000000123,
		Data:        []byte(`{"after":{"id":1,"name":"Ada"},"ts_ms":1710000000123}`),
	}
	row := map[string]interface{}{"id": 1, "name": "Ada"}
	columns := []string{"id", "name", "_cdc_op", "_cdc_ts", "_cdc_deleted", "_cdc_lsn"}

	args := clickhouseAppendArgs(columns, row, event)

	if args[2] != constant.OpUpdate.String() {
		t.Fatalf("_cdc_op = %#v", args[2])
	}
	if args[3] != time.UnixMilli(1710000000123).UTC() {
		t.Fatalf("_cdc_ts = %#v", args[3])
	}
	if args[4] != false {
		t.Fatalf("_cdc_deleted = %#v", args[4])
	}
	if args[5] != uint64(12345) {
		t.Fatalf("_cdc_lsn = %#v", args[5])
	}
}

func TestClickHouseDeleteAppendsTombstoneMetadata(t *testing.T) {
	event := &domain.Event{
		Op:          constant.OpDelete,
		LSN:         77,
		TimestampMS: 1710000000456,
		Data:        []byte(`{"before":{"id":1,"name":"Ada"},"ts_ms":1710000000456}`),
	}
	row := map[string]interface{}{"id": 1, "name": "Ada"}
	columns := []string{"id", "name", "_cdc_op", "_cdc_ts", "_cdc_deleted", "_cdc_lsn"}

	args := clickhouseAppendArgs(columns, row, event)

	if args[2] != constant.OpDelete.String() {
		t.Fatalf("_cdc_op = %#v", args[2])
	}
	if args[4] != true {
		t.Fatalf("_cdc_deleted = %#v", args[4])
	}
}

func TestBuildInsertSQLQuotesIdentifiers(t *testing.T) {
	query := buildInsertSQL("analytics.users", []string{"id", "name", "_cdc_op"})

	if !strings.HasPrefix(query, "INSERT INTO `analytics`.`users` (") {
		t.Fatalf("query = %s", query)
	}
	if !strings.Contains(query, "`id`") || !strings.Contains(query, "`_cdc_op`") {
		t.Fatalf("query does not quote columns: %s", query)
	}
}
