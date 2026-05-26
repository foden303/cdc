package discovery

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-mysql-org/go-mysql/driver"

	"github.com/foden/cdc/config"
	"github.com/foden/cdc/internal/core/ports"
)

type mysqlConnConfig struct {
	host     string
	port     int
	username string
	password string
	database string
}

func mysqlConfigFromSource(cfg *ports.SourceConfig) mysqlConnConfig {
	return mysqlConnConfig{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		database: cfg.Database,
	}
}

func testMySQLSourceConnection(ctx context.Context, cfg *ports.SourceConfig) error {
	return testMySQLConnection(ctx, mysqlConfigFromSource(cfg))
}

func testMySQLConnection(ctx context.Context, cfg mysqlConnConfig) error {
	db, err := openMySQLPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping mysql: %w", err)
	}
	return nil
}

func openMySQLPool(ctx context.Context, cfg mysqlConnConfig) (*sql.DB, error) {
	dsn := config.MySQLDSN(cfg.host, cfg.port, cfg.username, cfg.password, cfg.database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(connectionTimeout)
	return db, nil
}

func mysqlConfigFromSink(cfg *ports.SinkConfig) mysqlConnConfig {
	return mysqlConnConfig{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		database: cfg.Database,
	}
}

func testMySQLSinkConnection(ctx context.Context, cfg *ports.SinkConfig) error {
	return testMySQLConnection(ctx, mysqlConfigFromSink(cfg))
}

func openMySQLDB(cfg *ports.SourceConfig) (*sql.DB, error) {
	dsn := config.MySQLDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(connectionTimeout)
	return db, nil
}

func openMySQLSinkDB(cfg *ports.SinkConfig) (*sql.DB, error) {
	dsn := config.MySQLDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(connectionTimeout)
	return db, nil
}

func discoverMySQLSourceTables(ctx context.Context, cfg *ports.SourceConfig) ([]ports.TableInfo, error) {
	db, err := openMySQLDB(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tableRows, err := db.QueryContext(ctx, `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE'
		ORDER BY table_schema, table_name
	`, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to query mysql tables: %w", err)
	}
	defer tableRows.Close()

	var tableKeys []discoveredTable
	for tableRows.Next() {
		var tk discoveredTable
		if err := tableRows.Scan(&tk.schema, &tk.name); err != nil {
			return nil, fmt.Errorf("failed to scan mysql table row: %w", err)
		}
		tableKeys = append(tableKeys, tk)
	}
	if err := tableRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating mysql table rows: %w", err)
	}
	if len(tableKeys) == 0 {
		return []ports.TableInfo{}, nil
	}

	pkRows, err := db.QueryContext(ctx, `
		SELECT table_schema, table_name, column_name
		FROM information_schema.key_column_usage
		WHERE table_schema = ?
		  AND constraint_name = 'PRIMARY'
	`, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to query mysql primary keys: %w", err)
	}
	defer pkRows.Close()

	primaryKeys := make(map[columnMapKey]bool)
	for pkRows.Next() {
		var schema, table, column string
		if err := pkRows.Scan(&schema, &table, &column); err != nil {
			return nil, fmt.Errorf("failed to scan mysql primary key row: %w", err)
		}
		primaryKeys[newColumnMapKey(schema, table, column)] = true
	}
	if err := pkRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating mysql primary key rows: %w", err)
	}

	colRows, err := db.QueryContext(ctx, `
		SELECT table_schema, table_name, column_name, column_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = ?
		ORDER BY table_schema, table_name, ordinal_position
	`, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to query mysql columns: %w", err)
	}
	defer colRows.Close()

	columnsByTable := make(map[tableMapKey][]discoveredColumn)
	for colRows.Next() {
		var col discoveredColumn
		var isNullable string
		if err := colRows.Scan(&col.schema, &col.table, &col.name, &col.dataType, &isNullable); err != nil {
			return nil, fmt.Errorf("failed to scan mysql column row: %w", err)
		}
		col.isNullable = isNullable == "YES"
		key := newTableMapKey(col.schema, col.table)
		columnsByTable[key] = append(columnsByTable[key], col)
	}
	if err := colRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating mysql column rows: %w", err)
	}

	return assembleTables(tableKeys, columnsByTable, primaryKeys), nil
}

func discoverMySQLSinkTables(ctx context.Context, cfg *ports.SinkConfig) ([]ports.TableInfo, error) {
	sourceLike := &ports.SourceConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		Database: cfg.Database,
	}
	return discoverMySQLSourceTables(ctx, sourceLike)
}
