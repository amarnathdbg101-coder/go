package raft_test

import (
	"sync"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-raft"
)

// In-Memory cluster transport mock
type mockCluster struct {
	mu    sync.RWMutex
	nodes map[string]*raft.Node
}

func newMockCluster() *mockCluster {
	return &mockCluster{
		nodes: make(map[string]*raft.Node),
	}
}

func (c *mockCluster) register(id string, n *raft.Node) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodes[id] = n
}

func (c *mockCluster) SendRequestVote(target string, args *raft.RequestVoteArgs) (*raft.RequestVoteReply, error) {
	c.mu.RLock()
	node, exists := c.nodes[target]
	c.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	reply := &raft.RequestVoteReply{}
	node.HandleRequestVote(args, reply)
	return reply, nil
}

func (c *mockCluster) SendAppendEntries(target string, args *raft.AppendEntriesArgs) (*raft.AppendEntriesReply, error) {
	c.mu.RLock()
	node, exists := c.nodes[target]
	c.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	reply := &raft.AppendEntriesReply{}
	node.HandleAppendEntries(args, reply)
	return reply, nil
}

func TestRaft_LeaderElection(t *testing.T) {
	cluster := newMockCluster()
	peerIDs := []string{"node-1", "node-2", "node-3"}
	nodes := make([]*raft.Node, 3)

	for i, id := range peerIDs {
		var peers []string
		for _, p := range peerIDs {
			if p != id {
				peers = append(peers, p)
			}
		}

		applyCh := make(chan raft.ApplyMsg, 10)
		n, err := raft.New(
			id,
			peers,
			cluster,
			applyCh,
			raft.WithElectionTimeout(100*time.Millisecond),
			raft.WithHeartbeatInterval(30*time.Millisecond),
		)
		if err != nil {
			t.Fatalf("failed to create node %s: %v", id, err)
		}
		nodes[i] = n
		cluster.register(id, n)
	}

	defer func() {
		for _, n := range nodes {
			n.Stop()
		}
	}()

	// Wait for election to complete (polling up to 2 seconds)
	var leaderNode *raft.Node
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		leaders := 0
		for _, n := range nodes {
			_, isLeader := n.GetState()
			if isLeader {
				leaders++
				leaderNode = n
			}
		}
		if leaders == 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if leaderNode == nil {
		t.Fatal("expected a leader to be elected within deadline")
	}

	// Verify non-leader cannot propose
	for _, n := range nodes {
		if n != leaderNode {
			_, _, err := n.Propose([]byte("test_cmd"))
			if err != raft.ErrNotLeader {
				t.Fatalf("expected ErrNotLeader, got %v", err)
			}
		}
	}
}

func TestRaft_LogReplication(t *testing.T) {
	cluster := newMockCluster()
	peerIDs := []string{"node-A", "node-B", "node-C"}
	nodes := make([]*raft.Node, 3)
	applyChans := make([]chan raft.ApplyMsg, 3)

	for i, id := range peerIDs {
		var peers []string
		for _, p := range peerIDs {
			if p != id {
				peers = append(peers, p)
			}
		}

		applyChans[i] = make(chan raft.ApplyMsg, 10)
		n, _ := raft.New(
			id,
			peers,
			cluster,
			applyChans[i],
			raft.WithElectionTimeout(100*time.Millisecond),
			raft.WithHeartbeatInterval(30*time.Millisecond),
		)
		nodes[i] = n
		cluster.register(id, n)
	}

	defer func() {
		for _, n := range nodes {
			n.Stop()
		}
	}()

	// Wait for election
	var leader *raft.Node
	var leaderIdx int
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for i, n := range nodes {
			if _, isL := n.GetState(); isL {
				leader = n
				leaderIdx = i
				break
			}
		}
		if leader != nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if leader == nil {
		t.Fatal("no leader elected within deadline")
	}

	// Propose command to leader
	cmdPayload := []byte("set:balance:Amar=5000")
	idx, _, err := leader.Propose(cmdPayload)
	if err != nil {
		t.Fatalf("propose failed: %v", err)
	}
	if idx != 1 {
		t.Fatalf("expected log index 1, got %d", idx)
	}

	// Wait for replication and commit
	select {
	case msg := <-applyChans[leaderIdx]:
		if string(msg.Command) != string(cmdPayload) {
			t.Fatalf("applied command mismatch: expected %s, got %s", cmdPayload, msg.Command)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for log commit on leader applyCh")
	}
}

