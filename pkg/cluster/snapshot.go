package cluster

import (
	"compress/gzip"
	"encoding/json"
	"log/slog"

	"github.com/hashicorp/raft"
)

// TopicMeta holds metadata about a topic for snapshotting.
type TopicMeta struct {
	Name       string `json:"name"`
	Partitions int    `json:"partitions"`
}

// SnapshotData is the serialized form of the broker state.
type SnapshotData struct {
	Topics []TopicMeta `json:"topics"`
}

// BrokerSnapshot implements raft.FSMSnapshot.
type BrokerSnapshot struct {
	broker interface {
		// We only need topic metadata for snapshot
	}
	data            []byte
	compressEnabled bool
}

// Persist writes the snapshot to the given sink.
func (s *BrokerSnapshot) Persist(sink raft.SnapshotSink) error {
	if s.compressEnabled {
		// Write compressed snapshot
		gzWriter := gzip.NewWriter(sink)
		if _, err := gzWriter.Write(s.data); err != nil {
			sink.Cancel()
			gzWriter.Close()
			return err
		}
		if err := gzWriter.Close(); err != nil {
			sink.Cancel()
			return err
		}
	} else {
		// Write uncompressed snapshot
		if _, err := sink.Write(s.data); err != nil {
			sink.Cancel()
			return err
		}
	}
	return sink.Close()
}

// Release is called when the snapshot is no longer needed.
func (s *BrokerSnapshot) Release() {}

// CreateSnapshotData builds snapshot data from the broker.
// This is called during Snapshot() to serialize current state.
func CreateSnapshotData(topics map[string]int) ([]byte, error) {
	snap := SnapshotData{
		Topics: make([]TopicMeta, 0, len(topics)),
	}

	for name, partitions := range topics {
		snap.Topics = append(snap.Topics, TopicMeta{
			Name:       name,
			Partitions: partitions,
		})
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}

	slog.Debug("snapshot created", "topics", len(snap.Topics), "bytes", len(data))
	return data, nil
}
