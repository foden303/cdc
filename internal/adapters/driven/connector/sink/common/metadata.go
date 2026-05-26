package common

import (
	"fmt"
	"strings"
)

type TableMetadata struct {
	Schema      string
	Table       string
	Columns     []string
	PrimaryKeys []string
	UpsertSQL   string
	DeleteSQL   string
}

func PostgresTableKey(schema, table string) (string, TableMetadata, error) {
	return tableKey(schema, table, "public")
}

func MySQLTableKey(database, table string) (string, TableMetadata, error) {
	return tableKey(database, table, "")
}

func tableKey(schema, table, defaultSchema string) (string, TableMetadata, error) {
	schema = strings.TrimSpace(schema)
	table = strings.TrimSpace(table)

	if strings.Contains(table, ".") {
		parts := strings.SplitN(table, ".", 2)
		schema = strings.TrimSpace(parts[0])
		table = strings.TrimSpace(parts[1])
	}
	if schema == "" {
		schema = defaultSchema
	}
	if table == "" {
		return "", TableMetadata{}, fmt.Errorf("table name is required")
	}
	if schema == "" {
		return table, TableMetadata{Table: table}, nil
	}
	return schema + "." + table, TableMetadata{Schema: schema, Table: table}, nil
}
