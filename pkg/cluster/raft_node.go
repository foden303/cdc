package cluster

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/foden/cdc/pkg/config"
	"github.com/foden/cdc/pkg/queue"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

const (
	raftTimeout    = 10 * time.Second
	snapshotRetain = 2
)

// RaftNode wraps a hashicorp/raft instance with the broker FSM.
type RaftNode struct {
	raft        *raft.Raft
	fsm         *BrokerFSM
	config      *config.ClusterConfig
	proposalSem chan struct{} // Backpressure semaphore
}

// NewRaftNode creates and starts a Raft node.
func NewRaftNode(cfg *config.ClusterConfig, broker *queue.Broker) (*RaftNode, error) {
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create raft data dir: %w", err)
	}

	// Convert timeout from milliseconds to duration
	raftTimeout := time.Duration(cfg.RaftTimeoutMs) * time.Millisecond

	// Raft config
	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.NodeID)

	// Log store + Stable store (BoltDB)
	boltPath := filepath.Join(cfg.DataDir, "raft.db")
	boltStore, err := raftboltdb.NewBoltStore(boltPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create bolt store: %w", err)
	}

	// Snapshot store - use configurable retention count
	snapshotStore, err := raft.NewFileSnapshotStore(cfg.DataDir, cfg.SnapshotRetention, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}

	// Transport
	addr, err := net.ResolveTCPAddr("tcp", cfg.BindAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bind address: %w", err)
	}

	transport, err := raft.NewTCPTransport(cfg.BindAddr, addr, 3, raftTimeout, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP transport: %w", err)
	}

	// FSM
	fsm := NewBrokerFSM(broker)
	fsm.SetSnapshotCompression(cfg.SnapshotCompressionEnabled)

	// Create Raft instance
	r, err := raft.NewRaft(raftCfg, fsm, boltStore, boltStore, snapshotStore, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft: %w", err)
	}

	node := &RaftNode{
		raft:        r,
		fsm:         fsm,
		config:      cfg,
		proposalSem: make(chan struct{}, cfg.ProposalQueueSize),
	}

	// Bootstrap cluster if this is the first node
	if cfg.Bootstrap {
		servers := []raft.Server{
			{
				ID:      raft.ServerID(cfg.NodeID),
				Address: raft.ServerAddress(cfg.BindAddr),
			},
		}

		// Add configured peers
		for _, peer := range cfg.Peers {
			servers = append(servers, raft.Server{
				ID:      raft.ServerID(peer),
				Address: raft.ServerAddress(peer),
			})
		}

		future := r.BootstrapCluster(raft.Configuration{
			Servers: servers,
		})
		if err := future.Error(); err != nil && err != raft.ErrCantBootstrap {
			slog.Warn("raft bootstrap", "err", err)
		}
	}

	slog.Info("raft node started",
		"node_id", cfg.NodeID,
		"bind_addr", cfg.BindAddr,
		"bootstrap", cfg.Bootstrap,
	)

	return node, nil
}

// Propose submits a command to the Raft cluster.
// Only the leader can accept proposals; returns an error if not leader.
func (n *RaftNode) Propose(cmdType CommandType, payload any) error {
	if n.raft.State() != raft.Leader {
		leader := n.raft.Leader()
		return fmt.Errorf("not leader, current leader: %s", leader)
	}

	// Backpressure: check if proposal queue is full
	select {
	case n.proposalSem <- struct{}{}:
		// Acquired slot
	default:
		return fmt.Errorf("proposal queue full, cannot accept new proposals")
	}
	defer func() { <-n.proposalSem }()

	// Validate command before applying to Raft
	if err := ValidateCommand(cmdType, payload); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	data, err := EncodeCommand(cmdType, payload)
	if err != nil {
		return fmt.Errorf("failed to encode command: %w", err)
	}

	raftTimeout := time.Duration(n.config.RaftTimeoutMs) * time.Millisecond
	future := n.raft.Apply(data, raftTimeout)
	if err := future.Error(); err != nil {
		return fmt.Errorf("raft apply failed: %w", err)
	}

	// Check if the FSM returned an error
	resp := future.Response()
	if err, ok := resp.(error); ok {
		return err
	}

	return nil
}

// IsLeader returns true if this node is the current Raft leader.
func (n *RaftNode) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

// LeaderAddr returns the address of the current leader.
func (n *RaftNode) LeaderAddr() string {
	addr, _ := n.raft.LeaderWithID()
	return string(addr)
}

// AddVoter adds a new node to the Raft cluster.
func (n *RaftNode) AddVoter(nodeID, addr string) error {
	raftTimeout := time.Duration(n.config.RaftTimeoutMs) * time.Millisecond
	future := n.raft.AddVoter(
		raft.ServerID(nodeID),
		raft.ServerAddress(addr),
		0,
		raftTimeout,
	)
	return future.Error()
}

// RemoveServer removes a node from the Raft cluster.
func (n *RaftNode) RemoveServer(nodeID string) error {
	raftTimeout := time.Duration(n.config.RaftTimeoutMs) * time.Millisecond
	future := n.raft.RemoveServer(
		raft.ServerID(nodeID),
		0,
		raftTimeout,
	)
	return future.Error()
}

// Snapshot triggers a manual snapshot.
func (n *RaftNode) Snapshot() error {
	future := n.raft.Snapshot()
	return future.Error()
}

// Shutdown gracefully stops the Raft node.
func (n *RaftNode) Shutdown() error {
	future := n.raft.Shutdown()
	return future.Error()
}

// ShutdownWithTimeout gracefully stops the Raft node with a timeout.
func (n *RaftNode) ShutdownWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		future := n.raft.Shutdown()
		done <- future.Error()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// Stats returns Raft internal stats for diagnostics.
func (n *RaftNode) Stats() map[string]string {
	return n.raft.Stats()
}

// GetLeaderID returns the ID of the current leader
func (n *RaftNode) GetLeaderID() string {
	id, _ := n.raft.LeaderWithID()
	return string(id)
}

// GetFollowersCount returns the number of followers in the cluster
func (n *RaftNode) GetFollowersCount() int {
	cfg := n.raft.GetConfiguration()
	if cfg == nil {
		return 0
	}
	servers := cfg.Configuration().Servers
	if len(servers) <= 1 {
		return 0
	}
	return len(servers) - 1 // Excluding leader
}

// LogMetrics logs current Raft state for observability
func (n *RaftNode) LogMetrics() {
	stats := n.Stats()
	slog.Info("raft metrics",
		"state", n.raft.State().String(),
		"leader", n.GetLeaderID(),
		"last_log_index", stats["last_log_index"],
		"last_log_term", stats["last_log_term"],
		"commit_index", stats["commit_index"],
		"followers", n.GetFollowersCount(),
	)
}
