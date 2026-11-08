package hashring

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestHashRing_BasicGetAndDeterminism(t *testing.T) {
	ring := New(50, "node-1", "node-2", "node-3")

	node1, err := ring.Get("user:1001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node2, _ := ring.Get("user:1001")
	if node1 != node2 {
		t.Errorf("expected deterministic node assignment, got %s and %s", node1, node2)
	}

	if ring.Len() != 3 {
		t.Errorf("expected 3 nodes, got %d", ring.Len())
	}
}

func TestHashRing_MinimalKeyMigration(t *testing.T) {
	// Start with 4 nodes
	ring := New(100, "node-1", "node-2", "node-3", "node-4")

	const totalKeys = 1000
	initialAssignments := make(map[string]string)

	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key-%d", i)
		n, _ := ring.Get(key)
		initialAssignments[key] = n
	}

	// Remove node-4
	ring.Remove("node-4")

	migratedCount := 0
	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key-%d", i)
		newN, _ := ring.Get(key)
		oldN := initialAssignments[key]

		if newN != oldN {
			migratedCount++
			// Keys that were NOT on node-4 should mostly stay where they were!
			if oldN != "node-4" {
				t.Logf("key %s migrated from %s to %s", key, oldN, newN)
			}
		}
	}

	// With 4 nodes, roughly 25% (20% to 35%) should migrate when 1 node is removed
	migrationRatio := float64(migratedCount) / float64(totalKeys)
	if migrationRatio < 0.18 || migrationRatio > 0.35 {
		t.Errorf("expected ~25%% migration, got %.2f%% (%d/%d)", migrationRatio*100, migratedCount, totalKeys)
	}
}

func TestHashRing_GetNReplicas(t *testing.T) {
	ring := New(50, "redis-a", "redis-b", "redis-c")

	replicas, err := ring.GetN("session-key", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(replicas) != 2 {
		t.Fatalf("expected 2 distinct replicas, got %d", len(replicas))
	}
	if replicas[0] == replicas[1] {
		t.Errorf("expected distinct replica nodes, got duplicates: %v", replicas)
	}

	// Request more than available
	allReplicas, _ := ring.GetN("session-key", 5)
	if len(allReplicas) != 3 {
		t.Errorf("expected max available nodes (3), got %d", len(allReplicas))
	}
}

func TestHashRing_Empty(t *testing.T) {
	ring := New(50)
	_, err := ring.Get("key")
	if !errors.Is(err, ErrEmptyRing) {
		t.Errorf("expected ErrEmptyRing, got: %v", err)
	}
}

func TestHashRing_Concurrency(t *testing.T) {
	ring := New(50, "srv-1", "srv-2")
	var wg sync.WaitGroup
	const routines = 30

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%5 == 0 {
				ring.Add(fmt.Sprintf("srv-%d", id))
			} else {
				ring.Get(fmt.Sprintf("key-%d", id))
			}
		}(i)
	}

	wg.Wait()
}
