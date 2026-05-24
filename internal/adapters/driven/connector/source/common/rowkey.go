package common

import (
	"fmt"
	"sort"
	"strings"
)

// BestEffortRowKey returns a stable key for partitioning/dedup from a generic row map.
// Priority:
// 1) id / <table>_id / uuid / key
// 2) first non-null value in sorted column order
func BestEffortRowKey(table string, row map[string]any) string {
	if len(row) == 0 {
		return ""
	}

	lowerTable := strings.ToLower(table)
	preferredCols := []string{"id", lowerTable + "_id", "uuid", "key"}
	lowerValueByName := make(map[string]any, len(row))
	for k, v := range row {
		lowerValueByName[strings.ToLower(k)] = v
	}

	for _, col := range preferredCols {
		if v, ok := lowerValueByName[col]; ok && v != nil {
			return fmt.Sprintf("%s=%v", col, v)
		}
	}

	keys := make([]string, 0, len(lowerValueByName))
	for k := range lowerValueByName {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if v := lowerValueByName[k]; v != nil {
			return fmt.Sprintf("%s=%v", k, v)
		}
	}

	return ""
}
