package crdt_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-crdt"
)

func TestPNCounterConvergence(t *testing.T) {
	nodeA := crdt.NewPNCounter("pos_terminal_1")
	nodeB := crdt.NewPNCounter("pos_terminal_2")

	// Node A sells 10 items, refunds 2 items
	nodeA.Inc(10)
	nodeA.Dec(2)

	// Node B sells 5 items, refunds 1 item
	nodeB.Inc(5)
	nodeB.Dec(1)

	// Before merge
	if nodeA.Value() != 8 {
		t.Fatalf("expected nodeA value 8, got %d", nodeA.Value())
	}
	if nodeB.Value() != 4 {
		t.Fatalf("expected nodeB value 4, got %d", nodeB.Value())
	}

	// Two-way merge
	if err := nodeA.Merge(nodeB); err != nil {
		t.Fatalf("merge A <- B failed: %v", err)
	}
	if err := nodeB.Merge(nodeA); err != nil {
		t.Fatalf("merge B <- A failed: %v", err)
	}

	// After merge: (10 + 5) - (2 + 1) = 12
	if nodeA.Value() != 12 {
		t.Fatalf("expected converged nodeA value 12, got %d", nodeA.Value())
	}
	if nodeB.Value() != 12 {
		t.Fatalf("expected converged nodeB value 12, got %d", nodeB.Value())
	}
}

func TestPNCounterJSON(t *testing.T) {
	nodeA := crdt.NewPNCounter("device_x")
	nodeA.Inc(42)
	nodeA.Dec(7)

	data, err := json.Marshal(nodeA)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	nodeB := crdt.NewPNCounter("device_y")
	if err := json.Unmarshal(data, nodeB); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if nodeB.Value() != 35 {
		t.Fatalf("expected unmarshalled value 35, got %d", nodeB.Value())
	}
}

func TestLWWRegister(t *testing.T) {
	regA := crdt.NewLWWRegister("node_1", "initial_name")
	regB := crdt.NewLWWRegister("node_2", "initial_name")

	time.Sleep(5 * time.Millisecond)
	// Node B updates later
	regB.Set("updated_by_B")

	_ = regA.Merge(regB)
	if regA.Value() != "updated_by_B" {
		t.Fatalf("expected 'updated_by_B', got %s", regA.Value())
	}
}

func TestORSetAddRemove(t *testing.T) {
	setA := crdt.NewORSet[string]("node_A")
	setB := crdt.NewORSet[string]("node_B")

	// Both nodes independently add "Parle-G"
	setA.Add("Parle-G")
	setB.Add("Maggi")

	_ = setA.Merge(setB)
	_ = setB.Merge(setA)

	if !setA.Contains("Parle-G") || !setA.Contains("Maggi") {
		t.Fatalf("both items should be present after merge")
	}

	// Node A removes "Parle-G"
	setA.Remove("Parle-G")
	if setA.Contains("Parle-G") {
		t.Fatalf("Parle-G should be removed on node A")
	}

	// Sync again
	_ = setB.Merge(setA)
	if setB.Contains("Parle-G") {
		t.Fatalf("Parle-G should be removed on node B after sync")
	}

	// Add again (Add-Wins)
	setB.Add("Parle-G")
	_ = setA.Merge(setB)
	if !setA.Contains("Parle-G") {
		t.Fatalf("Parle-G should be resurrected when newly added")
	}
}

