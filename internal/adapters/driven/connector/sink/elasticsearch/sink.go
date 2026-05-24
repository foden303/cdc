package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"

	"github.com/foden/cdc/internal/adapters/driven/registry"
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/internal/core/ports"
	"github.com/foden/cdc/pkg/retry"
)

func init() {
	registry.RegisterSink(constant.SinkTypeElasticsearch.String(), func(cfg *ports.SinkConfig) (ports.Sink, error) {
		return New(cfg)
	})
}

// ElasticSink writes CDC events to Elasticsearch via the Bulk API.
type ElasticSink struct {
	client *elasticsearch.Client
	cfg    *ports.SinkConfig
}

// Internal structures for parsing Bulk API responses
type bulkResponse struct {
	Errors bool                        `json:"errors"`
	Items  []map[string]bulkItemResult `json:"items"`
}

type bulkItemResult struct {
	Index  string         `json:"_index"`
	ID     string         `json:"_id"`
	Status int            `json:"status"`
	Error  *bulkItemError `json:"error,omitempty"`
}

type bulkItemError struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

// New creates an ElasticSink and verifies connection.
func New(cfg *ports.SinkConfig) (*ElasticSink, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ElasticSink{client: client, cfg: cfg}, nil
}

// WriteBatch writes events to Elasticsearch using the Bulk API.
func (s *ElasticSink) WriteBatch(events []*domain.Event) error {
	var buf bytes.Buffer

	for _, event := range events {
		var node ast.Node
		var err error

		if event.Op == constant.OpDelete {
			node, err = sonic.Get(event.Data, "before")
		} else {
			node, err = sonic.Get(event.Data, "after")
		}

		if err != nil || !node.Exists() {
			continue
		}

		s.sanitizeNode(&node)

		docBytes, _ := node.MarshalJSON()
		index := s.indexName(event.InstanceID, event.Table)
		docID := extractIDFast(docBytes)

		if event.Op == constant.OpDelete {
			writeDeleteAction(&buf, index, docID)
		} else {
			writeIndexAction(&buf, index, docID, docBytes)
		}
	}

	if buf.Len() == 0 {
		return nil
	}

	data := buf.Bytes()
	var res *esapi.Response

	err := retry.Do(context.Background(), retry.Config{
		MaxAttempts: int(s.cfg.MaxRetries) + 1,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}, func() error {
		req := esapi.BulkRequest{Body: bytes.NewReader(data)}
		var reqErr error
		res, reqErr = req.Do(context.Background(), s.client)
		if reqErr != nil {
			return reqErr
		}
		if res.IsError() {
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			return fmt.Errorf("bulk request failed: status %d, body: %s", res.StatusCode, string(body))
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("bulk request failed after retries: %w", err)
	}
	defer res.Body.Close()

	// Parse item-level errors
	respBody, _ := io.ReadAll(res.Body)
	var bulkRes bulkResponse
	if err := json.Unmarshal(respBody, &bulkRes); err == nil && bulkRes.Errors {
		for _, item := range bulkRes.Items {
			for action, result := range item {
				if result.Error != nil {
					slog.Error("Bulk Item Error",
						"action", action,
						"id", result.ID,
						"reason", result.Error.Reason,
						"type", result.Error.Type)
				}
			}
		}
		return fmt.Errorf("bulk response contained errors")
	}

	slog.Debug("bulk write completed", "count", len(events))
	return nil
}

// Close is a no-op for the HTTP-based Elasticsearch client.
func (s *ElasticSink) Close() error {
	return nil
}

// InstanceID returns the sink instance ID.
func (s *ElasticSink) InstanceID() string {
	return s.cfg.InstanceID
}

// Type returns the type of the sink.
func (s *ElasticSink) Type() string {
	return constant.SinkTypeElasticsearch.String()
}

// newClient builds and pings the ES client.
func newClient(cfg *ports.SinkConfig) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.URL,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	if cfg.APIKey != "" {
		esCfg.APIKey = cfg.APIKey
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	res, err := client.Info()
	if err != nil {
		return nil, err
	}
	res.Body.Close()

	return client, nil
}

// indexName builds the target index name from prefix + instance + table.
func (s *ElasticSink) indexName(instanceID, table string) string {
	safeTable := strings.ReplaceAll(table, ".", "_")
	return fmt.Sprintf("%s%s_%s", s.cfg.IndexPrefix, instanceID, safeTable)
}

// extractIDFast tries to pull a document ID from the raw JSON payload quickly using sonic.Get.
func extractIDFast(doc []byte) string {
	for _, k := range []string{"id", "ID", "uuid", "uid", "guid"} {
		if node, err := sonic.Get(doc, k); err == nil {
			switch node.TypeSafe() {
			case ast.V_ARRAY:
				l, _ := node.Len()
				if l == 16 {
					b := make([]byte, 16)
					for i := 0; i < 16; i++ {
						v, _ := node.Index(i).Int64()
						b[i] = byte(v)
					}
					return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
				}
				var b []byte
				for i := 0; i < l; i++ {
					v, _ := node.Index(i).Int64()
					b = append(b, byte(v))
				}
				return fmt.Sprintf("%x", b)
			case ast.V_STRING:
				val, _ := node.String()
				return val
			default:
				val, _ := node.Raw()
				return strings.Trim(val, "\"")
			}
		}
	}
	return ""
}

// writeDeleteAction appends a bulk delete line.
func writeDeleteAction(buf *bytes.Buffer, index, docID string) {
	if docID == "" {
		return
	}
	buf.WriteString(fmt.Sprintf(`{"delete":{"_index":"%s","_id":"%s"}}`, index, docID))
	buf.WriteByte('\n')
}

// writeIndexAction appends a bulk index (upsert) line.
func writeIndexAction(buf *bytes.Buffer, index, docID string, doc []byte) {
	if docID != "" {
		buf.WriteString(fmt.Sprintf(`{"index":{"_index":"%s","_id":"%s"}}`, index, docID))
	} else {
		buf.WriteString(fmt.Sprintf(`{"index":{"_index":"%s"}}`, index))
	}
	buf.WriteByte('\n')
	buf.Write(doc)
	buf.WriteByte('\n')
}

// parseFlexTime parses various Postgres time formats.
func parseFlexTime(val string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05.999999-07",
		"2006-01-02 15:04:05.999999+00",
		"2006-01-02T15:04:05.999999Z",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	var lastErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, val)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

// sanitizeNode cleans up the AST node for ES mapping compatibility.
func (s *ElasticSink) sanitizeNode(node *ast.Node) {
	// 1. Fix 'metadata' object to string for keyword mapping compatibility
	meta := node.Get("metadata")
	if meta.Exists() && meta.TypeSafe() == ast.V_OBJECT {
		raw, _ := meta.Raw()
		if raw == "{}" {
			_, _ = node.Set("metadata", ast.NewString(""))
		} else {
			_, _ = node.Set("metadata", ast.NewString(raw))
		}
	}

	// 2. Convert time fields to Epoch Milliseconds
	if obj, err := node.Map(); err == nil {
		for key, val := range obj {
			strVal, isStr := val.(string)
			if !isStr {
				continue
			}

			if strings.HasSuffix(key, "_at") || strings.HasSuffix(key, "time") || key == "timestamp" {
				if t, err := parseFlexTime(strVal); err == nil {
					_, _ = node.Set(key, ast.NewAny(t.UnixMilli()))
				}
			}
		}
	}
}
