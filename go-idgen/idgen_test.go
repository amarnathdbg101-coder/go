package goidgen

import (
	"sync"
	"testing"
	"time"
)

func TestNode_Monotonic(t *testing.T) {
	node, err := NewNode(1)
	if err != nil {
		t.Fatalf("NewNode failed: %v", err)
	}

	id1, err := node.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	id2, err := node.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if id2 <= id1 {
		t.Fatalf("id2 (%d) should be strictly greater than id1 (%d)", id2, id1)
	}
}

func TestNode_ConcurrentNoDuplicates(t *testing.T) {
	node, _ := NewNode(42)

	const goroutines = 10
	const perGoroutine = 500
	total := goroutines * perGoroutine

	idChan := make(chan int64, total)
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				id, err := node.Generate()
				if err != nil {
					t.Errorf("generate error: %v", err)
					return
				}
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	seen := make(map[int64]struct{}, total)
	for id := range idChan {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate ID detected: %d", id)
		}
		seen[id] = struct{}{}
	}

	if len(seen) != total {
		t.Fatalf("expected %d unique IDs, got %d", total, len(seen))
	}
}

func TestNode_Deconstruct(t *testing.T) {
	nodeID := int64(99)
	node, _ := NewNode(nodeID)

	before := time.Now().UTC()
	id, err := node.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	after := time.Now().UTC()

	ts, nID, seq := node.Deconstruct(id)

	if nID != nodeID {
		t.Fatalf("expected nodeID %d, got %d", nodeID, nID)
	}
	if seq != 0 {
		t.Fatalf("expected sequence 0, got %d", seq)
	}
	if ts.Before(before.Add(-time.Second)) || ts.After(after.Add(time.Second)) {
		t.Fatalf("deconstructed timestamp %v out of range [%v, %v]", ts, before, after)
	}
}

func TestBase62_Roundtrip(t *testing.T) {
	node, _ := NewNode(5)
	id, _ := node.Generate()

	encoded := EncodeBase62(uint64(id))
	if len(encoded) == 0 {
		t.Fatal("encoded Base62 should not be empty")
	}

	decoded, err := DecodeBase62(encoded)
	if err != nil {
		t.Fatalf("DecodeBase62 failed: %v", err)
	}

	if decoded != uint64(id) {
		t.Fatalf("roundtrip mismatch: expected %d, got %d", id, decoded)
	}
}

func TestNode_InvalidNodeID(t *testing.T) {
	_, err := NewNode(1024) // 1024 exceeds 1023
	if err == nil {
		t.Fatal("expected error for nodeID > 1023, got nil")
	}

	_, err = NewNode(-1)
	if err == nil {
		t.Fatal("expected error for negative nodeID, got nil")
	}
}

func BenchmarkGenerate(b *testing.B) {
	node, _ := NewNode(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = node.Generate()
	}
}
