package btree_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-btree"
)

func TestBTree_BasicOps(t *testing.T) {
	tr, err := btree.New[int, string](btree.WithDegree(3))
	if err != nil {
		t.Fatalf("failed to create btree: %v", err)
	}

	if tr.Len() != 0 {
		t.Fatalf("expected len 0, got %d", tr.Len())
	}

	tr.Set(10, "ten")
	tr.Set(20, "twenty")
	tr.Set(5, "five")
	tr.Set(15, "fifteen")

	if tr.Len() != 4 {
		t.Fatalf("expected len 4, got %d", tr.Len())
	}

	// Get tests
	val, ok := tr.Get(10)
	if !ok || val != "ten" {
		t.Fatalf("expected ten, got %s (ok=%v)", val, ok)
	}

	_, ok = tr.Get(99)
	if ok {
		t.Fatal("key 99 should not exist")
	}

	// Update existing key
	old, updated := tr.Set(10, "TEN")
	if !updated || old != "ten" {
		t.Fatalf("expected update of ten, got old=%s (updated=%v)", old, updated)
	}
	val, _ = tr.Get(10)
	if val != "TEN" {
		t.Fatalf("expected TEN, got %s", val)
	}

	// Min / Max
	minK, minV, ok := tr.Min()
	if !ok || minK != 5 || minV != "five" {
		t.Fatalf("unexpected min: %d -> %s", minK, minV)
	}

	maxK, maxV, ok := tr.Max()
	if !ok || maxK != 20 || maxV != "twenty" {
		t.Fatalf("unexpected max: %d -> %s", maxK, maxV)
	}
}

func TestBTree_AscendAndRange(t *testing.T) {
	tr, _ := btree.New[int, int](btree.WithDegree(4))

	const total = 500
	for i := 0; i < total; i++ {
		tr.Set(i*2, i*20) // even keys: 0, 2, 4, ... 998
	}

	if tr.Len() != total {
		t.Fatalf("expected %d keys, got %d", total, tr.Len())
	}

	// Test Ascend order
	prev := -1
	count := 0
	tr.Ascend(func(k, v int) bool {
		if k <= prev {
			t.Fatalf("keys not strictly sorted: prev=%d, cur=%d", prev, k)
		}
		prev = k
		count++
		return true
	})

	if count != total {
		t.Fatalf("expected to ascend all %d keys, got %d", total, count)
	}

	// Test AscendRange [100, 200]
	var rangeKeys []int
	tr.AscendRange(100, 200, func(k, v int) bool {
		rangeKeys = append(rangeKeys, k)
		return true
	})

	if len(rangeKeys) != 51 { // 100, 102, ... 200 is 51 items
		t.Fatalf("expected 51 items in range [100, 200], got %d", len(rangeKeys))
	}
	if rangeKeys[0] != 100 || rangeKeys[len(rangeKeys)-1] != 200 {
		t.Fatalf("range boundaries invalid: first=%d, last=%d", rangeKeys[0], rangeKeys[len(rangeKeys)-1])
	}
}

func TestBTree_Delete(t *testing.T) {
	tr, _ := btree.New[int, string](btree.WithDegree(3))

	for i := 1; i <= 20; i++ {
		tr.Set(i, fmt.Sprintf("val_%d", i))
	}

	// Delete leaf key
	val, ok := tr.Delete(20)
	if !ok || val != "val_20" {
		t.Fatalf("failed to delete 20: %v (val=%s)", ok, val)
	}
	if tr.Len() != 19 {
		t.Fatalf("expected len 19, got %d", tr.Len())
	}
	_, found := tr.Get(20)
	if found {
		t.Fatal("key 20 should have been deleted")
	}

	// Delete internal key
	val, ok = tr.Delete(10)
	if !ok || val != "val_10" {
		t.Fatalf("failed to delete 10: %v (val=%s)", ok, val)
	}
	if tr.Len() != 18 {
		t.Fatalf("expected len 18, got %d", tr.Len())
	}
}

func TestBTree_StressAndConcurrency(t *testing.T) {
	tr, _ := btree.New[int, int](btree.WithDegree(16))
	var wg sync.WaitGroup

	// Insert 2000 elements concurrently
	writers := 10
	itemsPerWriter := 200

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(workerID)))
			for i := 0; i < itemsPerWriter; i++ {
				k := workerID*1000 + r.Intn(1000)
				tr.Set(k, k*2)
			}
		}(w)
	}

	// 5 concurrent readers
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_, _ = tr.Get(i * 10)
				_, _, _ = tr.Min()
			}
		}()
	}

	wg.Wait()

	// Verify all keys remain strictly sorted after concurrent insertions
	prev := -1
	tr.Ascend(func(k, v int) bool {
		if k <= prev {
			t.Fatalf("keys not sorted after concurrency: prev=%d, cur=%d", prev, k)
		}
		prev = k
		return true
	})
}

