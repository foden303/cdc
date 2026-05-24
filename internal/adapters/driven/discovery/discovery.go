package discovery

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/elastic/go-elasticsearch/v9"
	_ "github.com/go-mysql-org/go-mysql/driver"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/foden/cdc/config"
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/ports"
)

const connectionTimeout = 5 * time.Second

// Compile-time assertion that Service implements ports.Discovery.
var _ ports.Discovery = (*Service)(nil)

// Service provides connection testing and table discovery capabilities.
// It performs dynamic discovery at request time with no caching.
type Service struct{}

// NewService creates a new discovery Service.
func NewService() *Service {
	return &Service{}
}

// TestSourceConnection dials the source database and returns the round-trip latency in milliseconds.
func (s *Service) TestSourceConnection(ctx context.Context, cfg *ports.SourceConfig) (latencyMs int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	start := time.Now()

	switch cfg.Type {
	case constant.SourceTypePostgres.String():
		err = testPostgresSourceConnection(ctx, cfg)
	case constant.SourceTypeMySQL.String():
		err = testMySQLSourceConnection(ctx, cfg)
	default:
		err = fmt.Errorf("unsupported source type: %s", cfg.Type)
	}

	if err != nil {
		return 0, err
	}

	latencyMs = time.Since(start).Milliseconds()
	return latencyMs, nil
}

// TestSinkConnection dials the sink endpoint and returns the round-trip latency in milliseconds.
func (s *Service) TestSinkConnection(ctx context.Context, cfg *ports.SinkConfig) (latencyMs int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	start := time.Now()

	switch cfg.Type {
	case constant.SinkTypePostgres.String():
		err = testPostgresSinkConnection(ctx, cfg)
	case constant.SinkTypeElasticsearch.String():
		err = testElasticsearchConnection(ctx, cfg)
	case constant.SinkTypeClickhouse.String():
		err = testClickhouseConnection(ctx, cfg)
	default:
		err = fmt.Errorf("unsupported sink type: %s", cfg.Type)
	}

	if err != nil {
		return 0, err
	}

	latencyMs = time.Since(start).Milliseconds()
	return latencyMs, nil
}

// DiscoverSourceTables connects to the source database and returns table/column metadata.
func (s *Service) DiscoverSourceTables(ctx context.Context, cfg *ports.SourceConfig) ([]ports.TableInfo, error) {
	switch cfg.Type {
	case constant.SourceTypePostgres.String():
		return discoverPostgresSourceTables(ctx, cfg)
	case constant.SourceTypeMySQL.String():
		return discoverMySQLSourceTables(ctx, cfg)
	default:
		return []ports.TableInfo{}, nil
	}
}

// DiscoverSinkTables connects to the sink and returns table/column metadata.
func (s *Service) DiscoverSinkTables(ctx context.Context, cfg *ports.SinkConfig) ([]ports.TableInfo, error) {
	switch cfg.Type {
	case constant.SinkTypePostgres.String():
		return discoverPostgresSinkTables(ctx, cfg)
	case constant.SinkTypeClickhouse.String():
		return discoverClickhouseSinkTables(ctx, cfg)
	default:
		return []ports.TableInfo{}, nil
	}
}

// --- Source connection tests ---

func testPostgresSourceConnection(ctx context.Context, cfg *ports.SourceConfig) error {
	connStr := config.PostgresDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	return nil
}

func testMySQLSourceConnection(ctx context.Context, cfg *ports.SourceConfig) error {
	dsn := config.MySQLDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open mysql connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping mysql: %w", err)
	}

	return nil
}

// --- Sink connection tests ---

func testPostgresSinkConnection(ctx context.Context, cfg *ports.SinkConfig) error {
	connStr := config.PostgresDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	return nil
}

func testElasticsearchConnection(_ context.Context, cfg *ports.SinkConfig) error {
	esCfg := elasticsearch.Config{
		Addresses: cfg.URL,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			ResponseHeaderTimeout: connectionTimeout,
		},
	}
	if cfg.APIKey != "" {
		esCfg.APIKey = cfg.APIKey
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned error: %s", res.Status())
	}

	return nil
}

func testClickhouseConnection(ctx context.Context, cfg *ports.SinkConfig) error {
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
		DialTimeout: connectionTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to open clickhouse connection: %w", err)
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return nil
}

// --- Source table discovery ---

func discoverPostgresSourceTables(ctx context.Context, cfg *ports.SourceConfig) ([]ports.TableInfo, error) {
	connStr := config.PostgresDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pool.Close()

	return discoverPostgresTables(ctx, pool)
}

// --- Sink table discovery ---

func discoverPostgresSinkTables(ctx context.Context, cfg *ports.SinkConfig) ([]ports.TableInfo, error) {
	connStr := config.PostgresDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres sink: %w", err)
	}
	defer pool.Close()

	return discoverPostgresTables(ctx, pool)
}

// discoverPostgresTables is shared logic for discovering tables from a postgres connection.
func discoverPostgresTables(ctx context.Context, pool *pgxpool.Pool) ([]ports.TableInfo, error) {
	// 1. Discover tables
	tableRows, err := pool.Query(ctx, `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		  AND table_type = 'BASE TABLE'
		ORDER BY table_schema, table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer tableRows.Close()

	type tableKey struct {
		schema string
		name   string
	}
	var tableKeys []tableKey
	for tableRows.Next() {
		var tk tableKey
		if err := tableRows.Scan(&tk.schema, &tk.name); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}
		tableKeys = append(tableKeys, tk)
	}
	if err := tableRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	if len(tableKeys) == 0 {
		return []ports.TableInfo{}, nil
	}

	// 2. Discover primary keys
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
		return nil, fmt.Errorf("failed to query primary keys: %w", err)
	}
	defer pkRows.Close()

	pkSet := make(map[string]bool)
	for pkRows.Next() {
		var schema, table, column string
		if err := pkRows.Scan(&schema, &table, &column); err != nil {
			return nil, fmt.Errorf("failed to scan pk row: %w", err)
		}
		pkSet[schema+"."+table+"."+column] = true
	}
	if err := pkRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pk rows: %w", err)
	}

	// 3. Discover columns
	colRows, err := pool.Query(ctx, `
		SELECT table_schema, table_name, column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		ORDER BY table_schema, table_name, ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer colRows.Close()

	type colInfo struct {
		name       string
		dataType   string
		isNullable string
	}
	colMap := make(map[string][]colInfo)
	for colRows.Next() {
		var schema, table, colName, dataType, isNullable string
		if err := colRows.Scan(&schema, &table, &colName, &dataType, &isNullable); err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}
		key := schema + "." + table
		colMap[key] = append(colMap[key], colInfo{
			name:       colName,
			dataType:   dataType,
			isNullable: isNullable,
		})
	}
	if err := colRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column rows: %w", err)
	}

	// 4. Assemble results
	tables := make([]ports.TableInfo, 0, len(tableKeys))
	for _, tk := range tableKeys {
		key := tk.schema + "." + tk.name
		cols := colMap[key]
		columns := make([]ports.ColumnInfo, 0, len(cols))
		for _, c := range cols {
			columns = append(columns, ports.ColumnInfo{
				Name:         c.name,
				Type:         c.dataType,
				IsPrimaryKey: pkSet[tk.schema+"."+tk.name+"."+c.name],
				IsNullable:   c.isNullable == "YES",
			})
		}
		tables = append(tables, ports.TableInfo{
			Schema:  tk.schema,
			Name:    tk.name,
			Columns: columns,
		})
	}

	return tables, nil
}

func discoverMySQLSourceTables(ctx context.Context, cfg *ports.SourceConfig) ([]ports.TableInfo, error) {
	dsn := config.MySQLDSN(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}
	defer db.Close()

	// 1. Discover tables
	tableRows, err := db.QueryContext(ctx, `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE'
		ORDER BY table_schema, table_name
	`, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer tableRows.Close()

	type tableKey struct {
		schema string
		name   string
	}
	var tableKeys []tableKey
	for tableRows.Next() {
		var tk tableKey
		if err := tableRows.Scan(&tk.schema, &tk.name); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}
		tableKeys = append(tableKeys, tk)
	}
	if err := tableRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	if len(tableKeys) == 0 {
		return []ports.TableInfo{}, nil
	}

	// 2. Discover primary keys
	pkRows, err := db.QueryContext(ctx, `
		SELECT table_schema, table_name, column_name
		FROM information_schema.key_column_usage
		WHERE table_schema = ?
		  AND constraint_name = 'PRIMARY'
	`, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to query primary keys: %w", err)
	}
	defer pkRows.Close()

	pkSet := make(map[string]bool)
	for pkRows.Next() {
		var schema, table, column string
		if err := pkRows.Scan(&schema, &table, &column); err != nil {
			return nil, fmt.Errorf("failed to scan pk row: %w", err)
		}
		pkSet[schema+"."+table+"."+column] = true
	}
	if err := pkRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pk rows: %w", err)
	}

	// 3. Discover columns
	colRows, err := db.QueryContext(ctx, `
		SELECT table_schema, table_name, column_name, column_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = ?
		ORDER BY table_schema, table_name, ordinal_position
	`, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer colRows.Close()

	type colInfo struct {
		name       string
		dataType   string
		isNullable string
	}
	colMap := make(map[string][]colInfo)
	for colRows.Next() {
		var schema, table, colName, dataType, isNullable string
		if err := colRows.Scan(&schema, &table, &colName, &dataType, &isNullable); err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}
		key := schema + "." + table
		colMap[key] = append(colMap[key], colInfo{
			name:       colName,
			dataType:   dataType,
			isNullable: isNullable,
		})
	}
	if err := colRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column rows: %w", err)
	}

	// 4. Assemble results
	tables := make([]ports.TableInfo, 0, len(tableKeys))
	for _, tk := range tableKeys {
		key := tk.schema + "." + tk.name
		cols := colMap[key]
		columns := make([]ports.ColumnInfo, 0, len(cols))
		for _, c := range cols {
			columns = append(columns, ports.ColumnInfo{
				Name:         c.name,
				Type:         c.dataType,
				IsPrimaryKey: pkSet[tk.schema+"."+tk.name+"."+c.name],
				IsNullable:   c.isNullable == "YES",
			})
		}
		tables = append(tables, ports.TableInfo{
			Schema:  tk.schema,
			Name:    tk.name,
			Columns: columns,
		})
	}

	return tables, nil
}

func discoverClickhouseSinkTables(ctx context.Context, cfg *ports.SinkConfig) ([]ports.TableInfo, error) {
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
		DialTimeout: connectionTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}
	defer conn.Close()

	database := cfg.Database
	if database == "" {
		database = "default"
	}

	// 1. Discover tables from system.tables
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

	type tableKey struct {
		database string
		name     string
	}
	var tableKeys []tableKey
	for tableRows.Next() {
		var tk tableKey
		if err := tableRows.Scan(&tk.database, &tk.name); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}
		tableKeys = append(tableKeys, tk)
	}

	if len(tableKeys) == 0 {
		return []ports.TableInfo{}, nil
	}

	// 2. Discover columns from system.columns
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

	type colInfo struct {
		name         string
		dataType     string
		isPrimaryKey uint8
	}
	colMap := make(map[string][]colInfo)
	for colRows.Next() {
		var db, table, colName, dataType string
		var isPK uint8
		if err := colRows.Scan(&db, &table, &colName, &dataType, &isPK); err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}
		key := db + "." + table
		colMap[key] = append(colMap[key], colInfo{
			name:         colName,
			dataType:     dataType,
			isPrimaryKey: isPK,
		})
	}

	// 3. Assemble results
	tables := make([]ports.TableInfo, 0, len(tableKeys))
	for _, tk := range tableKeys {
		key := tk.database + "." + tk.name
		cols := colMap[key]
		columns := make([]ports.ColumnInfo, 0, len(cols))
		for _, c := range cols {
			columns = append(columns, ports.ColumnInfo{
				Name:         c.name,
				Type:         c.dataType,
				IsPrimaryKey: c.isPrimaryKey == 1,
				IsNullable:   false, // ClickHouse doesn't expose a simple nullable flag in system.columns
			})
		}
		tables = append(tables, ports.TableInfo{
			Schema:  tk.database,
			Name:    tk.name,
			Columns: columns,
		})
	}

	return tables, nil
}
