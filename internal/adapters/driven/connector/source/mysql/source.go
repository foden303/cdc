package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/go-mysql-org/go-mysql/schema"

	sourcecommon "github.com/foden/cdc/internal/adapters/driven/connector/source/common"
	"github.com/foden/cdc/internal/adapters/driven/metrics"
	"github.com/foden/cdc/internal/adapters/driven/registry"
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/internal/core/ports"
	coreruntime "github.com/foden/cdc/internal/core/runtime"
	"github.com/foden/cdc/pkg/retry"
	"github.com/foden/cdc/pkg/utils"
)

// mysqlTask represents a row change from binlog
type mysqlTask struct {
	op             constant.Op
	db             string
	table          *schema.Table
	before         []interface{}
	after          []interface{}
	lsn            uint64
	offset         string
	msgID          string
	ts             int64
	partitionCount int
}

// Pool for reusing maps to reduce GC pressure
var columnPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 32)
	},
}

func init() {
	registry.RegisterSource(constant.SourceTypeMySQL.String(), NewMySQLSource)
	registry.RegisterSource(constant.SourceTypeMariaDB.String(), NewMySQLSource)
}

func NewMySQLSource(cfg *ports.SourceConfig) (ports.Source, error) {
	return New(cfg)
}

// MySQLSource streams CDC events via MySQL binlog replication.
type MySQLSource struct {
	cfg    *ports.SourceConfig
	canal  *canal.Canal
	events chan<- *domain.Event
	stop   chan struct{}

	// Single worker for strict ordering
	taskChan chan *mysqlTask
	wg       sync.WaitGroup

	runtimeRegistry *coreruntime.Registry
	runtimeMetrics  *coreruntime.Metrics
}

// New creates a MySQLSource.
func New(cfg *ports.SourceConfig) (*MySQLSource, error) {
	return &MySQLSource{
		cfg:             cfg,
		stop:            make(chan struct{}),
		taskChan:        make(chan *mysqlTask, 8192),
		runtimeRegistry: coreruntime.DefaultRegistry(),
		runtimeMetrics:  coreruntime.DefaultMetrics(),
	}, nil
}

// Start initializes the canal and begins streaming events.
func (s *MySQLSource) Start(events chan<- *domain.Event, ackCh <-chan ports.SourceAck, initialOffset string) error {
	s.events = events

	// Start exactly ONE worker to guarantee 100% order
	s.wg.Add(1)
	go s.singleOrderedWorker()

	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	cfg.User = s.cfg.Username
	cfg.Password = s.cfg.Password
	cfg.Charset = "utf8mb4"
	if s.cfg.Type == constant.SourceTypeMariaDB.String() {
		cfg.Flavor = "mariadb"
	} else {
		cfg.Flavor = "mysql"
	}

	c, err := canal.NewCanal(cfg)
	if err != nil {
		return fmt.Errorf("failed to create canal: %w", err)
	}
	s.canal = c

	// Set event handler
	s.canal.SetEventHandler(&eventHandler{source: s})

	slog.Info("mysql replication starting", "host", s.cfg.Host, "instance", s.cfg.InstanceID, "offset", initialOffset)

	// Start ack loop
	go s.ackLoop(ackCh)

	// Run canal with auto-reconnect
	go func() {
		for {
			select {
			case <-s.stop:
				return
			default:
			}

			err := retry.Do(context.Background(), retry.SourceReconnectConfig(), func() error {
				select {
				case <-s.stop:
					return nil
				default:
				}

				var runErr error
				if initialOffset != "" {
					parts := strings.Split(initialOffset, ":")
					if len(parts) == 2 {
						pos, pErr := strconv.ParseUint(parts[1], 10, 32)
						if pErr == nil {
							mysqlPos := mysql.Position{Name: parts[0], Pos: uint32(pos)}
							runErr = s.canal.RunFrom(mysqlPos)
						} else {
							runErr = s.canal.Run()
						}
					} else {
						runErr = s.canal.Run()
					}
				} else {
					runErr = s.canal.Run()
				}

				if runErr != nil {
					if strings.Contains(runErr.Error(), "closed") {
						return nil // Graceful shutdown
					}
					slog.Error("mysql canal run failed, reconnecting",
						"err", runErr,
						"instance", s.cfg.InstanceID)

					// Update offset from last synced position for resume
					pos := s.canal.SyncedPosition()
					initialOffset = fmt.Sprintf("%s:%d", pos.Name, pos.Pos)
					return runErr
				}
				return nil
			})

			if err != nil {
				slog.Error("mysql source reconnect failed permanently", "err", err)
			}
			return
		}
	}()

	return nil
}

func (s *MySQLSource) Stop() error {
	slog.Info("stopping mysql source", "instance", s.cfg.InstanceID)
	close(s.stop)
	close(s.taskChan) // Signal worker to stop
	s.wg.Wait()       // Wait for processing to finish

	if s.canal != nil {
		s.canal.Close()
	}
	return nil
}

func (s *MySQLSource) InstanceID() string {
	return s.cfg.InstanceID
}

func (s *MySQLSource) SyncSourceTables(_ context.Context, tables []ports.SourceTableRef) error {
	slog.Info("mysql source tables reconciled", "instance", s.cfg.InstanceID, "tables", len(tables))
	return nil
}

func (s *MySQLSource) ackLoop(ackCh <-chan ports.SourceAck) {
	for {
		select {
		case <-s.stop:
			return
		case ack, ok := <-ackCh:
			if !ok {
				return
			}
			if ack.Offset != "" {
				slog.Debug("mysql source checkpoint acknowledged",
					"instance", s.cfg.InstanceID,
					"offset", ack.Offset,
					"lsn", ack.LSN)
			}
		}
	}
}

type eventHandler struct {
	canal.DummyEventHandler
	source *MySQLSource
}

func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	interest, active := h.source.runtimeRegistry.LookupTable(h.source.cfg.InstanceID, e.Table.Schema, e.Table.Name)
	if !active {
		return nil // No flow needs this table
	}
	partitionCount := int(interest.PartitionCount)

	changeIdx := 0
	for i := 0; i < len(e.Rows); {
		var before, after []interface{}
		op := constant.OpUpdate

		switch e.Action {
		case canal.InsertAction:
			op = constant.OpCreate
			after = e.Rows[i]
			i++
		case canal.DeleteAction:
			op = constant.OpDelete
			before = e.Rows[i]
			i++
		case canal.UpdateAction:
			op = constant.OpUpdate
			if i+1 < len(e.Rows) {
				before = e.Rows[i]
				after = e.Rows[i+1]
			}
			i += 2
		}

		lsn := uint64(e.Header.LogPos)
		pos := h.source.canal.SyncedPosition()
		offset := fmt.Sprintf("%s:%d", pos.Name, pos.Pos)
		msgID := fmt.Sprintf(
			"%s-mysql-%s-%d-%d-%d-%s.%s-%s",
			h.source.cfg.InstanceID,
			pos.Name,
			pos.Pos,
			e.Header.LogPos,
			changeIdx,
			e.Table.Schema,
			e.Table.Name,
			string(op),
		)
		changeIdx++

		h.source.taskChan <- &mysqlTask{
			op:             op,
			db:             e.Table.Schema,
			table:          e.Table,
			before:         before,
			after:          after,
			lsn:            lsn,
			offset:         offset,
			msgID:          msgID,
			ts:             time.Now().UnixMilli(),
			partitionCount: partitionCount,
		}
	}
	return nil
}

func (s *MySQLSource) singleOrderedWorker() {
	defer s.wg.Done()
	for task := range s.taskChan {
		s.processTask(task)
	}
}

func (s *MySQLSource) processTask(t *mysqlTask) {
	before := s.rowToMap(t.table, t.before)
	after := s.rowToMap(t.table, t.after)

	beforeData, _ := sonic.Marshal(before)
	afterData, _ := sonic.Marshal(after)

	// Recycle maps back to pool
	if before != nil {
		s.releaseMap(before)
	}
	if after != nil {
		s.releaseMap(after)
	}

	payload := domain.DebeziumPayload{
		Op:     t.op,
		Before: json.RawMessage(beforeData),
		After:  json.RawMessage(afterData),
		Source: domain.SourceMetadata{
			Version:   "1.0",
			Connector: "mysql",
			Name:      s.cfg.InstanceID,
			TsMs:      t.ts,
			Snapshot:  "false",
			DB:        t.db,
			Schema:    t.db,
			Table:     t.table.Name,
			LSN:       t.lsn,
		},
		TimestampMS: time.Now().UnixMilli(),
	}

	data, _ := sonic.Marshal(payload)

	// Use "cdc" as the default topic (derived from instance)
	topic := "cdc"

	// Calculate Partition ID based on Primary Key
	partitionID := s.calculatePartition(t)

	// Build Hierarchical Subject (5 levels): cdc.{instance_id}.{schema}.{table}.{partition_id}
	subject := fmt.Sprintf("%s.%s.%s.%s.%d", topic, s.cfg.InstanceID, t.db, t.table.Name, partitionID)

	ev := sourcecommon.BuildEvent(
		topic,
		subject,
		s.cfg.InstanceID,
		t.db,
		t.table.Name,
		t.op,
		t.lsn,
		t.offset,
		data,
		partitionID,
		t.msgID,
	)

	s.events <- ev
	if s.runtimeMetrics != nil {
		s.runtimeMetrics.RecordSourceProduced(s.cfg.InstanceID, t.db, t.table.Name, 1, t.ts)
	}
	metrics.EventsProducedTotal.WithLabelValues(s.cfg.InstanceID, "success").Inc()
}

// calculatePartition hashes the Primary Key of the row to determine the destination partition.
func (s *MySQLSource) calculatePartition(t *mysqlTask) int {
	partitionCount := t.partitionCount
	if partitionCount <= 0 {
		partitionCount = 4
	}

	var pkValues []string

	if t.table == nil || len(t.table.PKColumns) == 0 {
		return 0
	}

	// Extract values for PK columns
	// t.after is the row state after the change
	targetRow := t.after
	if t.op == constant.OpDelete {
		targetRow = t.before
	}

	if targetRow == nil {
		return 0
	}

	for _, pkIdx := range t.table.PKColumns {
		if pkIdx < len(targetRow) {
			pkValues = append(pkValues, formatPKValue(targetRow[pkIdx]))
		}
	}

	if len(pkValues) == 0 {
		return 0
	}

	// Use the configured partition count
	return utils.GeneratePartition(utils.CombineKeys(pkValues...), partitionCount)
}

func (s *MySQLSource) rowToMap(table *schema.Table, row []interface{}) map[string]interface{} {
	if row == nil {
		return nil
	}

	obj := columnPool.Get().(map[string]interface{})
	for i, val := range row {
		name := table.Columns[i].Name
		obj[name] = formatValue(val)
	}
	return obj
}

func (s *MySQLSource) releaseMap(m map[string]interface{}) {
	for k := range m {
		delete(m, k)
	}
	columnPool.Put(m)
}

func formatPKValue(val interface{}) string {
	switch v := val.(type) {
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatValue(val interface{}) interface{} {
	switch v := val.(type) {
	case []byte:
		return string(v)
	default:
		return v
	}
}

func (h *eventHandler) OnRotate(header *replication.EventHeader, rotateEvent *replication.RotateEvent) error {
	return nil
}

func (h *eventHandler) OnDDL(header *replication.EventHeader, nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	return nil
}

func (h *eventHandler) String() string {
	return "MySQLCDCEventHandler"
}
