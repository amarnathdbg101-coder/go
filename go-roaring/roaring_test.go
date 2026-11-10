package roaring_test

import (
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-roaring"
)

func TestBitmap_BasicOps(t *testing.T) {
	bm := roaring.New()

	if bm.Cardinality() != 0 {
		t.Fatalf("expected cardinality 0, got %d", bm.Cardinality())
	}

	bm.Add(10)
	bm.Add(20)
	bm.Add(30)

	if bm.Cardinality() != 3 {
		t.Fatalf("expected cardinality 3, got %d", bm.Cardinality())
	}

	if !bm.Contains(10) || !bm.Contains(20) || !bm.Contains(30) {
		t.Fatal("expected bitmap to contain 10, 20, 30")
	}

	if bm.Contains(40) {
		t.Fatal("did not expect bitmap to contain 40")
	}

	bm.Remove(20)
	if bm.Contains(20) {
		t.Fatal("expected 20 to be removed")
	}
	if bm.Cardinality() != 2 {
		t.Fatalf("expected cardinality 2, got %d", bm.Cardinality())
	}

	arr := bm.ToArray()
	if len(arr) != 2 || arr[0] != 10 || arr[1] != 30 {
		t.Fatalf("unexpected slice output: %v", arr)
	}
}

func TestBitmap_ContainerUpgradeDowngrade(t *testing.T) {
	bm := roaring.New()

	// Insert 4100 items into the same high chunk (high = 0)
	// Threshold is 4096, so it must upgrade to bitmapContainer
	var i uint32
	for i = 0; i < 4100; i++ {
		bm.Add(i)
	}

	if bm.Cardinality() != 4100 {
		t.Fatalf("expected cardinality 4100, got %d", bm.Cardinality())
	}

	for i = 0; i < 4100; i++ {
		if !bm.Contains(i) {
			t.Fatalf("missing element %d after container upgrade", i)
		}
	}

	// Remove 100 items down to 4000 (below threshold 4096), triggering downgrade
	for i = 4000; i < 4100; i++ {
		bm.Remove(i)
	}

	if bm.Cardinality() != 4000 {
		t.Fatalf("expected cardinality 4000, got %d", bm.Cardinality())
	}

	for i = 4000; i < 4100; i++ {
		if bm.Contains(i) {
			t.Fatalf("element %d should have been removed", i)
		}
	}
}

func TestBitmap_SetOperations(t *testing.T) {
	bm1 := roaring.New()
	bm2 := roaring.New()

	// bm1: {1, 2, 3, 100000}
	bm1.AddMany(1, 2, 3, 100000)
	// bm2: {2, 3, 4, 100000, 200000}
	bm2.AddMany(2, 3, 4, 100000, 200000)

	// Intersection (And)
	andBm := bm1.And(bm2)
	expectedAnd := []uint32{2, 3, 100000}
	andSlice := andBm.ToArray()
	if len(andSlice) != len(expectedAnd) {
		t.Fatalf("expected And len %d, got %d", len(expectedAnd), len(andSlice))
	}
	for i, v := range expectedAnd {
		if andSlice[i] != v {
			t.Fatalf("And mismatch at index %d: expected %d, got %d", i, v, andSlice[i])
		}
	}

	// Union (Or)
	orBm := bm1.Or(bm2)
	expectedOr := []uint32{1, 2, 3, 4, 100000, 200000}
	orSlice := orBm.ToArray()
	if len(orSlice) != len(expectedOr) {
		t.Fatalf("expected Or len %d, got %d", len(expectedOr), len(orSlice))
	}
	for i, v := range expectedOr {
		if orSlice[i] != v {
			t.Fatalf("Or mismatch at index %d: expected %d, got %d", i, v, orSlice[i])
		}
	}

	// Difference (AndNot: bm1 \ bm2) -> {1}
	diffBm := bm1.AndNot(bm2)
	diffSlice := diffBm.ToArray()
	if len(diffSlice) != 1 || diffSlice[0] != 1 {
		t.Fatalf("expected diff [1], got %v", diffSlice)
	}

	// Clone test
	cloneBm := bm1.Clone()
	if cloneBm.Cardinality() != bm1.Cardinality() {
		t.Fatalf("expected clone cardinality %d, got %d", bm1.Cardinality(), cloneBm.Cardinality())
	}
	bm1.Add(999)
	if cloneBm.Contains(999) {
		t.Fatal("modifying original should not affect clone")
	}
}

func TestBitmap_Concurrency(t *testing.T) {
	bm := roaring.New()
	var wg sync.WaitGroup

	// 10 concurrent writers
	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			offset := uint32(workerID * 1000)
			for i := uint32(0); i < 100; i++ {
				bm.Add(offset + i)
			}
		}(w)
	}

	// 5 concurrent readers
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_ = bm.Cardinality()
				_ = bm.Contains(500)
			}
		}()
	}

	wg.Wait()

	if bm.Cardinality() != 1000 {
		t.Fatalf("expected 1000 items after concurrent insertions, got %d", bm.Cardinality())
	}
}

