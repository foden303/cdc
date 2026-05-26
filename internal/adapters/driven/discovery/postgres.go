package discovery

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/foden/cdc/config"
	"github.com/foden/cdc/internal/core/ports"
)

type postgresConnConfig struct {
	host     string
	port     int
	username string
	password string
	database string
}

func postgresConfigFromSource(cfg *ports.SourceConfig) postgresConnConfig {
	return postgresConnConfig{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		database: cfg.Database,
	}
}

func postgresConfigFromSink(cfg *ports.SinkConfig) postgresConnConfig {
	return postgresConnConfig{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		database: cfg.Database,
	}
}

func testPostgresConnection(ctx context.Context, cfg postgresConnConfig) error {
	pool, err := openPostgresPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}
	return nil
}

func openPostgresPool(ctx context.Context, cfg postgresConnConfig) (*pgxpool.Pool, error) {
	connStr := config.PostgresDSN(cfg.host, cfg.port, cfg.username, cfg.password, cfg.database)
	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid postgres connection string: %w", err)
	}
	poolCfg.MinConns = 0
	poolCfg.MaxConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}
	return pool, nil
}

func discoverPostgresSourceTables(ctx context.Context, cfg *ports.SourceConfig) ([]ports.TableInfo, error) {
	pool, err := openPostgresPool(ctx, postgresConfigFromSource(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres source: %w", err)
	}
	defer pool.Close()

	return discoverPostgresTables(ctx, pool)
}

func discoverPostgresSinkTables(ctx context.Context, cfg *ports.SinkConfig) ([]ports.TableInfo, error) {
	pool, err := openPostgresPool(ctx, postgresConfigFromSink(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres sink: %w", err)
	}
	defer pool.Close()

	return discoverPostgresTables(ctx, pool)
}

func discoverPostgresTables(ctx context.Context, pool *pgxpool.Pool) ([]ports.TableInfo, error) {
	tableRows, err := pool.Query(ctx, `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		  AND table_type = 'BASE TABLE'
		ORDER BY table_schema, table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query postgres tables: %w", err)
	}
	defer tableRows.Close()

	var tableKeys []discoveredTable
	for tableRows.Next() {
		var tk discoveredTable
		if err := tableRows.Scan(&tk.schema, &tk.name); err != nil {
			return nil, fmt.Errorf("failed to scan postgres table row: %w", err)
		}
		tableKeys = append(tableKeys, tk)
	}
	if err := tableRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating postgres table rows: %w", err)
	}
	if len(tableKeys) == 0 {
		return []ports.TableInfo{}, nil
	}

	pkRows, err := pool.Query(ctx, `
		SELECT kcu.table_schema, kcu.table_name, kcu.column_name
		FROM information_schema.key_column_usage kcu
		JOIN pg_constraint pgc
		  ON pgc.conname = kcu.constraint_name
		  AND pgc.connamespace = (
		    SELECT oid FROM pg_namespace WHERE nspname = kcu.table_schema
		  )
		WHERE pgc.contype = 'p'
		  AND kcu.table_schema NOT IN ('pg_catalog', 'information_schema')
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query postgres primary keys: %w", err)
	}
	defer pkRows.Close()

	primaryKeys := make(map[columnMapKey]bool)
	for pkRows.Next() {
		var schema, table, column string
		if err := pkRows.Scan(&schema, &table, &column); err != nil {
			return nil, fmt.Errorf("failed to scan postgres primary key row: %w", err)
		}
		primaryKeys[newColumnMapKey(schema, table, column)] = true
	}
	if err := pkRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating postgres primary key rows: %w", err)
	}

	colRows, err := pool.Query(ctx, `
		SELECT table_schema, table_name, column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		ORDER BY table_schema, table_name, ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query postgres columns: %w", err)
	}
	defer colRows.Close()

	columnsByTable := make(map[tableMapKey][]discoveredColumn)
	for colRows.Next() {
		var col discoveredColumn
		var isNullable string
		if err := colRows.Scan(&col.schema, &col.table, &col.name, &col.dataType, &isNullable); err != nil {
			return nil, fmt.Errorf("failed to scan postgres column row: %w", err)
		}
		col.isNullable = isNullable == "YES"
		key := newTableMapKey(col.schema, col.table)
		columnsByTable[key] = append(columnsByTable[key], col)
	}
	if err := colRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating postgres column rows: %w", err)
	}

	return assembleTables(tableKeys, columnsByTable, primaryKeys), nil
}
