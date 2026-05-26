package common

import (
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
)

var DefaultKeyCandidates = []string{"id", "ID", "uuid", "uid", "guid"}

func RowBytes(event *domain.Event) ([]byte, bool, error) {
	if event == nil || len(event.Data) == 0 {
		return nil, false, nil
	}
	field := "after"
	if event.Op == constant.OpDelete {
		field = "before"
	}
	node, err := sonic.Get(event.Data, field)
	if err != nil {
		return nil, false, fmt.Errorf("get %s row: %w", field, err)
	}
	if !node.Exists() {
		return nil, false, nil
	}
	data, err := node.MarshalJSON()
	if err != nil {
		return nil, false, fmt.Errorf("marshal %s row: %w", field, err)
	}
	return data, true, nil
}

func RowMap(event *domain.Event) (map[string]interface{}, bool, error) {
	data, ok, err := RowBytes(event)
	if err != nil || !ok {
		return nil, ok, err
	}
	var row map[string]interface{}
	if err := sonic.Unmarshal(data, &row); err != nil {
		return nil, false, fmt.Errorf("unmarshal row: %w", err)
	}
	return row, true, nil
}

func ExtractKeyValue(row map[string]interface{}, candidates ...string) (string, interface{}, bool) {
	if len(candidates) == 0 {
		candidates = DefaultKeyCandidates
	}
	for _, key := range candidates {
		value, ok := row[key]
		if ok && value != nil && value != "" {
			return key, value, true
		}
	}
	return "", nil, false
}

func ExtractKeyString(row map[string]interface{}, candidates ...string) (string, bool) {
	_, value, ok := ExtractKeyValue(row, candidates...)
	if !ok {
		return "", false
	}
	switch v := value.(type) {
	case string:
		return v, true
	case json.Number:
		return v.String(), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}
