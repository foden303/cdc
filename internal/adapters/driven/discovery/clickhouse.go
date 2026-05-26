package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/foden/cdc/internal/core/ports"
)

const (
	defaultClickhouseDatabase = "default"
	defaultClickhouseAddress  = "127.0.0.1:9000"
)

func testClickhouseConnection(ctx context.Context, cfg *ports.SinkConfig) error {
	conn, err := openClickhouseConnection(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping clickhouse: %w", err)
	}
	return nil
}

func discoverClickhouseSinkTables(ctx context.Context, cfg *ports.SinkConfig) ([]ports.TableInfo, error) {
	conn, err := openClickhouseConnection(cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	database := cfg.Database
	if database == "" {
		database = defaultClickhouseDatabase
	}

	tableRows, err := conn.Query(ctx, `
		SELECT database, name
		FROM system.tables
		WHERE database = ?
		  AND engine NOT IN ('View', 'MaterializedView', 'LiveView')
		ORDER BY database, name
	`, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query clickhouse tables: %w", err)
	}
	defer tableRows.Close()

	var tableKeys []discoveredTable
	for tableRows.Next() {
		var tk discoveredTable
		if err := tableRows.Scan(&tk.schema, &tk.name); err != nil {
			return nil, fmt.Errorf("failed to scan clickhouse table row: %w", err)
		}
		tableKeys = append(tableKeys, tk)
	}
	if err := tableRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clickhouse table rows: %w", err)
	}
	if len(tableKeys) == 0 {
		return []ports.TableInfo{}, nil
	}

	colRows, err := conn.Query(ctx, `
		SELECT database, table, name, type, is_in_primary_key
		FROM system.columns
		WHERE database = ?
		ORDER BY database, table, position
	`, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query clickhouse columns: %w", err)
	}
	defer colRows.Close()

	columnsByTable := make(map[tableMapKey][]discoveredColumn)
	for colRows.Next() {
		var col discoveredColumn
		var isPrimaryKey uint8
		if err := colRows.Scan(&col.schema, &col.table, &col.name, &col.dataType, &isPrimaryKey); err != nil {
			return nil, fmt.Errorf("failed to scan clickhouse column row: %w", err)
		}
		col.isPrimaryKey = isPrimaryKey == 1
		col.isNullable = isClickhouseNullableType(col.dataType)
		key := newTableMapKey(col.schema, col.table)
		columnsByTable[key] = append(columnsByTable[key], col)
	}
	if err := colRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clickhouse column rows: %w", err)
	}

	return assembleTables(tableKeys, columnsByTable, nil), nil
}

func openClickhouseConnection(cfg *ports.SinkConfig) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{clickhouseAddr(cfg)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout:     connectionTimeout,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: connectionTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}
	return conn, nil
}

func clickhouseAddr(cfg *ports.SinkConfig) string {
	if cfg.Host != "" && cfg.Port > 0 {
		return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}
	if cfg.Host != "" {
		return cfg.Host
	}
	if len(cfg.URL) > 0 {
		return cfg.URL[0]
	}
	return defaultClickhouseAddress
}

func isClickhouseNullableType(dataType string) bool {
	normalized := strings.TrimSpace(dataType)
	if strings.HasPrefix(normalized, "Nullable(") {
		return true
	}
	if strings.HasPrefix(normalized, "LowCardinality(") && strings.HasSuffix(normalized, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(normalized, "LowCardinality("), ")")
		return isClickhouseNullableType(inner)
	}
	return false
}
