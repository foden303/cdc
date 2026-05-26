package postgres

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/foden/cdc/config"
	"github.com/foden/cdc/internal/core/ports"
	"github.com/foden/cdc/pkg/utils"
	"github.com/jackc/pgx/v5"
)

func normalizeSourceTableRef(ref ports.SourceTableRef) ports.SourceTableRef {
	schema := strings.TrimSpace(ref.Schema)
	table := strings.TrimSpace(ref.Table)
	if schema == "" {
		schema = "public"
	}
	return ports.SourceTableRef{Schema: schema, Table: table}
}

func quoteQualifiedTable(ref ports.SourceTableRef) string {
	ref = normalizeSourceTableRef(ref)
	return fmt.Sprintf("%s.%s",
		utils.QuoteIdentifierDoubleQuote(ref.Schema),
		utils.QuoteIdentifierDoubleQuote(ref.Table),
	)
}

func createPublicationSQL(pubName string, tables []ports.SourceTableRef) string {
	quotedPub := utils.QuoteIdentifierDoubleQuote(pubName)
	tables = dedupeTables(tables)
	if len(tables) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tables))
	for _, table := range tables {
		parts = append(parts, quoteQualifiedTable(table))
	}
	return fmt.Sprintf("CREATE PUBLICATION %s FOR TABLE %s", quotedPub, strings.Join(parts, ", "))
}

func dedupeTables(tables []ports.SourceTableRef) []ports.SourceTableRef {
	seen := make(map[string]ports.SourceTableRef)
	for _, table := range tables {
		table = normalizeSourceTableRef(table)
		if table.Table == "" {
			continue
		}
		seen[table.Schema+"."+table.Table] = table
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]ports.SourceTableRef, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result
}

func tableSet(tables []ports.SourceTableRef) map[string]bool {
	result := make(map[string]bool, len(tables))
	for _, table := range dedupeTables(tables) {
		result[table.Schema+"."+table.Table] = true
	}
	return result
}

func currentPublicationTables(ctx context.Context, conn *pgx.Conn, pubName string) (map[string]ports.SourceTableRef, error) {
	rows, err := conn.Query(ctx, `
		SELECT schemaname, tablename
		FROM pg_publication_tables
		WHERE pubname = $1
	`, pubName)
	if err != nil {
		return nil, fmt.Errorf("list publication tables: %w", err)
	}
	defer rows.Close()

	result := make(map[string]ports.SourceTableRef)
	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil {
			return nil, fmt.Errorf("scan publication table: %w", err)
		}
		ref := normalizeSourceTableRef(ports.SourceTableRef{Schema: schema, Table: table})
		result[ref.Schema+"."+ref.Table] = ref
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate publication tables: %w", err)
	}
	return result, nil
}

func (p *PostgresSource) SyncSourceTables(ctx context.Context, tables []ports.SourceTableRef) error {
	connStr := config.PostgresDSN(p.cfg.Host, p.cfg.Port, p.cfg.Username, p.cfg.Password, p.cfg.Database)
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("publication sync connection failed: %w", err)
	}
	defer conn.Close(ctx)

	pubName := p.publicationName()
	quotedPub := utils.QuoteIdentifierDoubleQuote(pubName)
	desired := dedupeTables(tables)
	if len(desired) == 0 {
		if _, err := conn.Exec(ctx, fmt.Sprintf("DROP PUBLICATION IF EXISTS %s", quotedPub)); err != nil {
			return fmt.Errorf("drop empty publication: %w", err)
		}
		return nil
	}

	var exists bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_publication WHERE pubname = $1)", pubName).Scan(&exists); err != nil {
		return fmt.Errorf("check publication exists: %w", err)
	}
	if !exists {
		if _, err := conn.Exec(ctx, createPublicationSQL(pubName, desired)); err != nil {
			return fmt.Errorf("create publication: %w", err)
		}
		return nil
	}

	var allTables bool
	if err := conn.QueryRow(ctx, "SELECT puballtables FROM pg_publication WHERE pubname = $1", pubName).Scan(&allTables); err != nil {
		return fmt.Errorf("check publication scope: %w", err)
	}
	if allTables {
		if _, err := conn.Exec(ctx, fmt.Sprintf("DROP PUBLICATION %s", quotedPub)); err != nil {
			return fmt.Errorf("drop all-tables publication: %w", err)
		}
		if _, err := conn.Exec(ctx, createPublicationSQL(pubName, desired)); err != nil {
			return fmt.Errorf("create table-scoped publication: %w", err)
		}
		return nil
	}

	current, err := currentPublicationTables(ctx, conn, pubName)
	if err != nil {
		return err
	}
	for _, table := range desired {
		key := table.Schema + "." + table.Table
		if _, ok := current[key]; ok {
			continue
		}
		if _, err := conn.Exec(ctx, fmt.Sprintf("ALTER PUBLICATION %s ADD TABLE %s", quotedPub, quoteQualifiedTable(table))); err != nil {
			return fmt.Errorf("add publication table %s: %w", key, err)
		}
	}

	desiredSet := tableSet(desired)
	for key, table := range current {
		if desiredSet[key] {
			continue
		}
		if _, err := conn.Exec(ctx, fmt.Sprintf("ALTER PUBLICATION %s DROP TABLE %s", quotedPub, quoteQualifiedTable(table))); err != nil {
			return fmt.Errorf("drop publication table %s: %w", key, err)
		}
	}
	return nil
}
