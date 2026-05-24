package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/foden/cdc/config"
	"github.com/foden/cdc/internal/adapters/driven/registry"
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/internal/core/ports"
)

func init() {
	registry.RegisterSink(constant.SinkTypePostgres.String(), func(cfg *ports.SinkConfig) (ports.Sink, error) {
		return New(cfg)
	})
}

// PostgresSink writes CDC events to another PostgreSQL database.
type PostgresSink struct {
	pool *pgxpool.Pool
	cfg  *ports.SinkConfig

	// tableCache tracks tables we've already ensured exist in this session
	tableCache sync.Map
}

// New creates a new PostgresSink instance.
func New(cfg *ports.SinkConfig) (*PostgresSink, error) {
	ctx := context.Background()
	connStr := config.PostgresDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return &PostgresSink{pool: pool, cfg: cfg}, nil
}

// WriteBatch writes events to PostgreSQL in a single transaction.
func (s *PostgresSink) WriteBatch(events []*domain.Event) error {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, event := range events {
		var node ast.Node
		var parseErr error

		if event.Op == constant.OpDelete {
			node, parseErr = sonic.Get(event.Data, "before")
		} else {
			node, parseErr = sonic.Get(event.Data, "after")
		}

		if parseErr != nil || !node.Exists() {
			continue
		}

		docBytes, err := node.MarshalJSON()
		if err != nil {
			continue
		}

		var data map[string]interface{}
		if err := sonic.Unmarshal(docBytes, &data); err != nil {
			slog.Error("unmarshal data failed in postgres sink", "err", err)
			continue
		}

		tableName := event.Table

		// Use "id" as default primary key (detection from data)
		pk := "id"

		pkValue, ok := data[pk]
		if !ok && event.Op != constant.OpCreate {
			slog.Warn("primary key not found in data", "pk", pk, "table", tableName)
			continue
		}

		// Self-healing: Ensure table exists
		if err := s.ensureTable(ctx, tableName, data, pk); err != nil {
			return fmt.Errorf("failed to ensure table %s: %w", tableName, err)
		}

		switch event.Op {
		case constant.OpDelete:
			query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", tableName, pk)
			if _, err := tx.Exec(ctx, query, pkValue); err != nil {
				return fmt.Errorf("delete failed: %w", err)
			}
		case constant.OpCreate, constant.OpUpdate, constant.OpSnapshot:
			if err := s.upsertTx(ctx, tx, tableName, pk, data); err != nil {
				return fmt.Errorf("upsert failed: %w", err)
			}
		default:
			slog.Warn("unknown operation type", "op", event.Op)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}
	return nil
}

func (s *PostgresSink) upsertTx(ctx context.Context, tx pgx.Tx, table string, pk string, data map[string]interface{}) error {
	cols := make([]string, 0, len(data))
	vals := make([]interface{}, 0, len(data))
	placeholders := make([]string, 0, len(data))
	updates := make([]string, 0, len(data)-1)

	i := 1
	for k, v := range data {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		if k != pk {
			updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", k, k))
		}
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		pk,
		strings.Join(updates, ", "),
	)

	_, err := tx.Exec(ctx, query, vals...)
	return err
}

func (s *PostgresSink) ensureTable(ctx context.Context, table string, data map[string]interface{}, pk string) error {
	if _, ok := s.tableCache.Load(table); ok {
		return nil
	}

	var exists bool
	checkQuery := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)"
	if err := s.pool.QueryRow(ctx, checkQuery, table).Scan(&exists); err != nil {
		return err
	}

	if exists {
		s.tableCache.Store(table, true)
		return nil
	}

	slog.Info("self-healing: creating table", "table", table, "pk", pk)
	cols := make([]string, 0, len(data))
	for k, v := range data {
		pgType := s.inferPGType(v)
		colDef := fmt.Sprintf("%s %s", k, pgType)
		if k == pk {
			colDef += " PRIMARY KEY"
		}
		cols = append(cols, colDef)
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", table, strings.Join(cols, ", "))
	if _, err := s.pool.Exec(ctx, createSQL); err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	s.tableCache.Store(table, true)
	return nil
}

func (s *PostgresSink) inferPGType(v interface{}) string {
	switch v.(type) {
	case bool:
		return "BOOLEAN"
	case int, int32, int64:
		return "BIGINT"
	case float32, float64:
		return "DOUBLE PRECISION"
	case string:
		return "TEXT"
	case map[string]interface{}, []interface{}:
		return "JSONB"
	default:
		return "TEXT"
	}
}

// Close closes the connection pool.
func (s *PostgresSink) Close() error {
	s.pool.Close()
	return nil
}

// InstanceID returns the sink instance ID.
func (s *PostgresSink) InstanceID() string {
	return s.cfg.InstanceID
}

// Type returns the sink type.
func (s *PostgresSink) Type() string {
	return constant.SinkTypePostgres.String()
}
