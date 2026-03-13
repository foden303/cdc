package cluster

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/foden/cdc/pkg/queue"
	"github.com/hashicorp/raft"
)

// BrokerFSM implements raft.FSM for replicating broker state.
type BrokerFSM struct {
	mu                      sync.RWMutex
	broker                  *queue.Broker
	snapshotCompressEnabled bool
}

// NewBrokerFSM creates a new FSM backed by the given broker.
func NewBrokerFSM(broker *queue.Broker) *BrokerFSM {
	return &BrokerFSM{
		broker: broker,
	}
}

// SetSnapshotCompression enables/disables snapshot compression
func (f *BrokerFSM) SetSnapshotCompression(enabled bool) {
	f.snapshotCompressEnabled = enabled
}

// Broker returns the underlying broker.
func (f *BrokerFSM) Broker() *queue.Broker {
	return f.broker
}

// Apply is called once a Raft log entry is committed.
// It applies the command to the local broker.
func (f *BrokerFSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		slog.Error("failed to unmarshal raft command", "err", err)
		return err
	}

	switch cmd.Type {
	case CmdProduce:
		return f.applyProduce(cmd.Data)
	case CmdCreateTopic:
		return f.applyCreateTopic(cmd.Data)
	default:
		return fmt.Errorf("unknown command type: %d", cmd.Type)
	}
}

func (f *BrokerFSM) applyProduce(data json.RawMessage) interface{} {
	var cmd ProduceCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return err
	}

	t := f.broker.Topic(cmd.Topic)
	if t == nil {
		return fmt.Errorf("topic %s not found", cmd.Topic)
	}

	msg := &queue.Message{
		Key:       cmd.Key,
		Value:     cmd.Value,
		Timestamp: cmd.Timestamp,
	}

	offset, err := t.Partition(cmd.Key).Produce(msg)
	if err != nil {
		return err
	}
	return offset
}

func (f *BrokerFSM) applyCreateTopic(data json.RawMessage) interface{} {
	var cmd CreateTopicCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return err
	}

	return f.broker.CreateTopic(cmd.Name, cmd.Partitions)
}

// Snapshot returns an FSMSnapshot for snapshotting.
func (f *BrokerFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	snap := &BrokerSnapshot{
		broker:          f.broker,
		compressEnabled: f.snapshotCompressEnabled,
	}

	return snap, nil
}

// Restore restores the FSM from a snapshot.
func (f *BrokerFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	// Try to decompress if it's gzip compressed
	var decoder *json.Decoder
	var reader io.ReadCloser

	// Attempt to detect and decompress gzip
	gzReader, err := gzip.NewReader(rc)
	if err == nil {
		// Successfully created gzip reader - use it
		decoder = json.NewDecoder(gzReader)
		reader = gzReader
	} else {
		// Not gzip - use raw reader
		decoder = json.NewDecoder(rc)
		reader = nil
	}

	var snap SnapshotData
	if err := decoder.Decode(&snap); err != nil {
		if reader != nil {
			reader.Close()
		}
		return fmt.Errorf("failed to decode snapshot: %w", err)
	}

	if reader != nil {
		reader.Close()
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Recreate topics from snapshot metadata
	for _, t := range snap.Topics {
		if err := f.broker.CreateTopic(t.Name, t.Partitions); err != nil {
			slog.Error("failed to restore topic", "name", t.Name, "err", err)
		}
	}

	slog.Info("FSM restored from snapshot", "topics", len(snap.Topics))
	return nil
}
