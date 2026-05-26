package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/foden/cdc/config"
	sinkcommon "github.com/foden/cdc/internal/adapters/driven/connector/sink/common"
	"github.com/foden/cdc/internal/adapters/driven/registry"
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/internal/core/ports"
	"github.com/foden/cdc/pkg/utils"
)

func init() {
	registry.RegisterSink(constant.SinkTypePostgres.String(), func(cfg *ports.SinkConfig) (ports.Sink, error) {
		return New(cfg)
	})
}

// PostgresSink writes CDC events to another PostgreSQL database.
type PostgresSink struct {
	pool          *pgxpool.Pool
	cfg           *ports.SinkConfig
	metadataCache sync.Map
	loadMetadata  func(context.Context, string, string) (sinkcommon.TableMetadata, error)
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

	sink := &PostgresSink{pool: pool, cfg: cfg}
	sink.loadMetadata = sink.loadTableMetadata
	return sink, nil
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
		data, ok, err := sinkcommon.RowMap(event)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		tableName := event.Table
		meta, err := s.metadataForTable(ctx, event.Schema, tableName)
		if err != nil {
			return err
		}

		pkValues, err := primaryKeyValues(data, meta.PrimaryKeys)
		if err != nil {
			return err
		}

		switch event.Op {
		case constant.OpDelete:
			if _, err := tx.Exec(ctx, meta.DeleteSQL, pkValues...); err != nil {
				return fmt.Errorf("delete failed: %w", err)
			}
		case constant.OpCreate, constant.OpUpdate, constant.OpSnapshot:
			values := valuesForColumns(data, meta.Columns)
			if _, err := tx.Exec(ctx, meta.UpsertSQL, values...); err != nil {
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

func (s *PostgresSink) metadataForTable(ctx context.Context, schema, table string) (sinkcommon.TableMetadata, error) {
	key, base, err := sinkcommon.PostgresTableKey(schema, table)
	if err != nil {
		return sinkcommon.TableMetadata{}, err
	}
	if cached, ok := s.metadataCache.Load(key); ok {
		return cached.(sinkcommon.TableMetadata), nil
	}
	loader := s.loadMetadata
	if loader == nil {
		loader = s.loadTableMetadata
	}
	meta, err := loader(ctx, base.Schema, base.Table)
	if err != nil {
		return sinkcommon.TableMetadata{}, err
	}
	meta.Schema = base.Schema
	meta.Table = base.Table
	if len(meta.Columns) == 0 {
		return sinkcommon.TableMetadata{}, fmt.Errorf("postgres sink table %s has no columns or does not exist", key)
	}
	if len(meta.PrimaryKeys) == 0 {
		return sinkcommon.TableMetadata{}, fmt.Errorf("postgres sink table %s has no primary key", key)
	}
	qualifiedTable := key
	meta.UpsertSQL = buildUpsertSQLForColumns(qualifiedTable, meta.PrimaryKeys, meta.Columns)
	meta.DeleteSQL = buildDeleteSQL(qualifiedTable, meta.PrimaryKeys)
	actual, _ := s.metadataCache.LoadOrStore(key, meta)
	return actual.(sinkcommon.TableMetadata), nil
}

func (s *PostgresSink) loadTableMetadata(ctx context.Context, schema, table string) (sinkcommon.TableMetadata, error) {
	columns, err := s.queryColumns(ctx, schema, table)
	if err != nil {
		return sinkcommon.TableMetadata{}, err
	}
	primaryKeys, err := s.queryPrimaryKeys(ctx, schema, table)
	if err != nil {
		return sinkcommon.TableMetadata{}, err
	}
	return sinkcommon.TableMetadata{Schema: schema, Table: table, Columns: columns, PrimaryKeys: primaryKeys}, nil
}

func (s *PostgresSink) queryColumns(ctx context.Context, schema, table string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = $1 AND table_name = $2
ORDER BY ordinal_position`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("query postgres columns for %s.%s: %w", schema, table, err)
	}
	defer rows.Close()

	columns := make([]string, 0)
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, fmt.Errorf("scan postgres column for %s.%s: %w", schema, table, err)
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read postgres columns for %s.%s: %w", schema, table, err)
	}
	return columns, nil
}

func (s *PostgresSink) queryPrimaryKeys(ctx context.Context, schema, table string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
SELECT kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
 AND tc.table_schema = kcu.table_schema
 AND tc.table_name = kcu.table_name
WHERE tc.constraint_type = 'PRIMARY KEY'
  AND tc.table_schema = $1
  AND tc.table_name = $2
ORDER BY kcu.ordinal_position`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("query postgres primary keys for %s.%s: %w", schema, table, err)
	}
	defer rows.Close()

	primaryKeys := make([]string, 0)
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, fmt.Errorf("scan postgres primary key for %s.%s: %w", schema, table, err)
		}
		primaryKeys = append(primaryKeys, column)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read postgres primary keys for %s.%s: %w", schema, table, err)
	}
	return primaryKeys, nil
}

func buildUpsertSQLForColumns(table string, primaryKeys []string, cols []string) string {
	pkSet := makeStringSet(primaryKeys)
	placeholders := make([]string, 0, len(cols))
	updates := make([]string, 0, len(cols)-1)
	for i, col := range cols {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		if !pkSet[col] {
			quoted := utils.QuoteIdentifierDoubleQuote(col)
			updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", quoted, quoted))
		}
	}
	if len(updates) == 0 {
		quotedPK := utils.QuoteIdentifierDoubleQuote(primaryKeys[0])
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", quotedPK, quotedPK))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		utils.QuoteIdentifierDoubleQuote(table),
		quotePostgresIdentifiers(cols),
		strings.Join(placeholders, ", "),
		quotePostgresIdentifiers(primaryKeys),
		strings.Join(updates, ", "),
	)

	return query
}

func quotePostgresIdentifiers(cols []string) string {
	quoted := make([]string, 0, len(cols))
	for _, col := range cols {
		quoted = append(quoted, utils.QuoteIdentifierDoubleQuote(col))
	}
	return strings.Join(quoted, ", ")
}

func buildDeleteSQL(table string, primaryKeys []string) string {
	clauses := make([]string, 0, len(primaryKeys))
	for i, pk := range primaryKeys {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", utils.QuoteIdentifierDoubleQuote(pk), i+1))
	}
	return fmt.Sprintf("DELETE FROM %s WHERE %s", utils.QuoteIdentifierDoubleQuote(table), strings.Join(clauses, " AND "))
}

func makeStringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func primaryKeyValues(row map[string]interface{}, primaryKeys []string) ([]interface{}, error) {
	values := make([]interface{}, 0, len(primaryKeys))
	for _, key := range primaryKeys {
		value, ok := row[key]
		if !ok || value == nil || value == "" {
			return nil, fmt.Errorf("missing primary key column %q", key)
		}
		values = append(values, value)
	}
	return values, nil
}

func valuesForColumns(row map[string]interface{}, columns []string) []interface{} {
	values := make([]interface{}, 0, len(columns))
	for _, column := range columns {
		values = append(values, row[column])
	}
	return values
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
