package discovery

import "github.com/foden/cdc/internal/core/ports"

type discoveredTable struct {
	schema string
	name   string
}

type discoveredColumn struct {
	schema       string
	table        string
	name         string
	dataType     string
	isPrimaryKey bool
	isNullable   bool
}

type tableMapKey struct {
	schema string
	table  string
}

type columnMapKey struct {
	schema string
	table  string
	column string
}

func newTableMapKey(schema, table string) tableMapKey {
	return tableMapKey{schema: schema, table: table}
}

func newColumnMapKey(schema, table, column string) columnMapKey {
	return columnMapKey{schema: schema, table: table, column: column}
}

func assembleTables(
	tableKeys []discoveredTable,
	columnsByTable map[tableMapKey][]discoveredColumn,
	primaryKeys map[columnMapKey]bool,
) []ports.TableInfo {
	if len(tableKeys) == 0 {
		return []ports.TableInfo{}
	}

	tables := make([]ports.TableInfo, 0, len(tableKeys))
	for _, tk := range tableKeys {
		cols := columnsByTable[newTableMapKey(tk.schema, tk.name)]
		columns := make([]ports.ColumnInfo, 0, len(cols))
		for _, col := range cols {
			columns = append(columns, ports.ColumnInfo{
				Name:         col.name,
				Type:         col.dataType,
				IsPrimaryKey: col.isPrimaryKey || primaryKeys[newColumnMapKey(tk.schema, tk.name, col.name)],
				IsNullable:   col.isNullable,
			})
		}
		tables = append(tables, ports.TableInfo{
			Schema:  tk.schema,
			Name:    tk.name,
			Columns: columns,
		})
	}

	return tables
}
