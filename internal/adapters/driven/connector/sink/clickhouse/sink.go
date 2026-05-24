package clickhouse

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"

	"github.com/foden/cdc/internal/adapters/driven/registry"
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/internal/core/ports"
)

func init() {
	registry.RegisterSink(constant.SinkTypeClickhouse.String(), func(cfg *ports.SinkConfig) (ports.Sink, error) {
		return New(cfg)
	})
}

// ClickhouseSink writes CDC events to ClickHouse.
type ClickhouseSink struct {
	conn clickhouse.Conn
	cfg  *ports.SinkConfig
}

// New creates a ClickhouseSink and verifies connection.
func New(cfg *ports.SinkConfig) (*ClickhouseSink, error) {
	addr := cfg.Host
	if cfg.Port > 0 {
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	} else if len(cfg.URL) > 0 {
		addr = cfg.URL[0]
	}

	if addr == "" {
		addr = "127.0.0.1:9000"
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Debug: true,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout: time.Second * 30,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return &ClickhouseSink{conn: conn, cfg: cfg}, nil
}

// WriteBatch writes events to ClickHouse grouped by table.
func (s *ClickhouseSink) WriteBatch(events []*domain.Event) error {
	// Group events by table as ClickHouse Bulk Insert is per table
	tableEvents := make(map[string][]*domain.Event)
	for _, event := range events {
		tableEvents[event.Table] = append(tableEvents[event.Table], event)
	}

	ctx := context.Background()
	for tableName, evts := range tableEvents {
		if err := s.writeTable(ctx, tableName, evts); err != nil {
			slog.Error("Clickhouse write table failed", "table", tableName, "error", err)
		}
	}
	return nil
}

func (s *ClickhouseSink) writeTable(ctx context.Context, tableName string, events []*domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Use the first event to determine columns.
	var firstNode ast.Node
	var err error
	if events[0].Op == constant.OpDelete {
		firstNode, err = sonic.Get(events[0].Data, "before")
	} else {
		firstNode, err = sonic.Get(events[0].Data, "after")
	}

	if err != nil || !firstNode.Exists() {
		return fmt.Errorf("failed to get first event node: %w", err)
	}

	firstData, _ := firstNode.MarshalJSON()
	var firstMap map[string]interface{}
	if err := sonic.Unmarshal(firstData, &firstMap); err != nil {
		return fmt.Errorf("failed to unmarshal first event: %w", err)
	}

	columns := make([]string, 0, len(firstMap))
	for k := range firstMap {
		columns = append(columns, k)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s)", tableName, strings.Join(columns, ", "))
	batch, err := s.conn.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, event := range events {
		var node ast.Node
		if event.Op == constant.OpDelete {
			node, err = sonic.Get(event.Data, "before")
		} else {
			node, err = sonic.Get(event.Data, "after")
		}

		if err != nil || !node.Exists() {
			continue
		}

		data, _ := node.MarshalJSON()

		var m map[string]interface{}
		_ = sonic.Unmarshal(data, &m)

		args := make([]interface{}, len(columns))
		for i, col := range columns {
			args[i] = m[col]
		}

		if err := batch.Append(args...); err != nil {
			slog.Error("Clickhouse append failed", "error", err)
		}
	}

	return batch.Send()
}

// Close closes the ClickHouse connection.
func (s *ClickhouseSink) Close() error {
	return s.conn.Close()
}

// InstanceID returns the sink instance ID.
func (s *ClickhouseSink) InstanceID() string {
	return s.cfg.InstanceID
}

// Type returns the sink type.
func (s *ClickhouseSink) Type() string {
	return constant.SinkTypeClickhouse.String()
}
