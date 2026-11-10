package rbtree_test

import (
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-rbtree"
)

func TestRBT_BasicOps(t *testing.T) {
	tr := rbtree.New[int, string]()

	if tr.Len() != 0 {
		t.Fatalf("expected len 0, got %d", tr.Len())
	}

	tr.Insert(10, "ten")
	tr.Insert(20, "twenty")
	tr.Insert(5, "five")
	tr.Insert(15, "fifteen")

	if tr.Len() != 4 {
		t.Fatalf("expected len 4, got %d", tr.Len())
	}

	val, ok := tr.Get(10)
	if !ok || val != "ten" {
		t.Fatalf("expected ten, got %s (ok=%v)", val, ok)
	}

	// Update existing key
	old, updated := tr.Insert(10, "TEN")
	if !updated || old != "ten" {
		t.Fatalf("expected update of ten, got old=%s (updated=%v)", old, updated)
	}
	val, _ = tr.Get(10)
	if val != "TEN" {
		t.Fatalf("expected TEN, got %s", val)
	}

	// Min & Max
	minK, minV, ok := tr.Min()
	if !ok || minK != 5 || minV != "five" {
		t.Fatalf("unexpected min: %d -> %s", minK, minV)
	}
	maxK, maxV, ok := tr.Max()
	if !ok || maxK != 20 || maxV != "twenty" {
		t.Fatalf("unexpected max: %d -> %s", maxK, maxV)
	}
}

func TestRBT_FloorAndCeiling(t *testing.T) {
	tr := rbtree.New[int, int]()

	tr.Insert(10, 10)
	tr.Insert(20, 20)
	tr.Insert(30, 30)

	// Floor(15) -> 10
	fK, _, ok := tr.Floor(15)
	if !ok || fK != 10 {
		t.Fatalf("floor(15) expected 10, got %d", fK)
	}

	// Ceiling(15) -> 20
	cK, _, ok := tr.Ceiling(15)
	if !ok || cK != 20 {
		t.Fatalf("ceiling(15) expected 20, got %d", cK)
	}

	// Floor(5) -> none
	_, _, ok = tr.Floor(5)
	if ok {
		t.Fatal("floor(5) should not exist")
	}
}

func TestRBT_SequentialInsertBalancing(t *testing.T) {
	tr := rbtree.New[int, int]()

	// Inserting sequentially into unbalanced BST creates linked list of depth 1000.
	// LLRB guarantees balanced tree where height <= 2 * log2(N).
	const n = 1000
	for i := 1; i <= n; i++ {
		tr.Insert(i, i*10)
	}

	if tr.Len() != n {
		t.Fatalf("expected %d elements, got %d", n, tr.Len())
	}

	// Verify all elements are strictly in ascending order
	prev := 0
	count := 0
	tr.Ascend(func(k, v int) bool {
		if k <= prev {
			t.Fatalf("tree not strictly sorted: prev=%d, cur=%d", prev, k)
		}
		prev = k
		count++
		return true
	})

	if count != n {
		t.Fatalf("expected count %d, got %d", n, count)
	}
}

func TestRBT_AscendRange(t *testing.T) {
	tr := rbtree.New[int, int]()

	for i := 0; i < 100; i++ {
		tr.Insert(i, i)
	}

	var rangeKeys []int
	tr.AscendRange(20, 25, func(k, v int) bool {
		rangeKeys = append(rangeKeys, k)
		return true
	})

	if len(rangeKeys) != 6 {
		t.Fatalf("expected 6 keys in [20, 25], got %d", len(rangeKeys))
	}
	if rangeKeys[0] != 20 || rangeKeys[5] != 25 {
		t.Fatalf("range boundaries invalid: %v", rangeKeys)
	}
}

func TestRBT_Concurrency(t *testing.T) {
	tr := rbtree.New[int, int]()
	var wg sync.WaitGroup

	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				tr.Insert(workerID*1000+i, i)
				_, _ = tr.Get(workerID*1000 + i)
			}
		}(w)
	}

	wg.Wait()
}

