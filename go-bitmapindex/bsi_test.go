package bitmapindex_test

import (
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-bitmapindex"
)

func TestBSI_SetAndGet(t *testing.T) {
	idx, err := bitmapindex.New(16) // 16-bit integers: 0 to 65535
	if err != nil {
		t.Fatalf("failed to create BSI: %v", err)
	}

	values := map[uint32]uint64{
		101: 500,
		102: 1250,
		103: 80,
		104: 0,
	}

	for r, v := range values {
		if err := idx.Set(r, v); err != nil {
			t.Fatalf("failed to set row %d: %v", r, err)
		}
	}

	if idx.Count() != 4 {
		t.Fatalf("expected 4 rows, got %d", idx.Count())
	}

	// Verify lookups
	for r, expected := range values {
		val, ok := idx.Get(r)
		if !ok || val != expected {
			t.Fatalf("row %d: expected %d, got %d (ok=%v)", r, expected, val, ok)
		}
	}

	// Non-existent row
	_, ok := idx.Get(999)
	if ok {
		t.Fatal("expected row 999 to not exist")
	}
}

func TestBSI_SumAndAvg(t *testing.T) {
	idx, _ := bitmapindex.New(16)

	// Insert prices for orders: 100, 200, 300, 400 (Total = 1000)
	_ = idx.Set(1, 100)
	_ = idx.Set(2, 200)
	_ = idx.Set(3, 300)
	_ = idx.Set(4, 400)

	totalSum := idx.Sum(nil)
	if totalSum != 1000 {
		t.Fatalf("expected total sum 1000, got %d", totalSum)
	}

	avg := idx.Avg(nil)
	if avg != 250.0 {
		t.Fatalf("expected avg 250.0, got %f", avg)
	}

	// Filtered sum for rows [1, 3] -> 100 + 300 = 400
	filteredSum := idx.Sum([]uint32{1, 3})
	if filteredSum != 400 {
		t.Fatalf("expected filtered sum 400, got %d", filteredSum)
	}
}

func TestBSI_FilterRange(t *testing.T) {
	idx, _ := bitmapindex.New(16)

	_ = idx.Set(10, 50)
	_ = idx.Set(11, 150)
	_ = idx.Set(12, 250)
	_ = idx.Set(13, 350)
	_ = idx.Set(14, 450)

	// Filter rows where value is between 100 and 300 (should return rows 11 and 12)
	matches, err := idx.FilterRange(100, 300)
	if err != nil {
		t.Fatalf("filter range failed: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0] != 11 || matches[1] != 12 {
		t.Fatalf("unexpected matching rows: %v", matches)
	}
}

func TestBSI_BitWidthOverflow(t *testing.T) {
	idx, _ := bitmapindex.New(8) // 8-bit integers: 0 to 255

	err := idx.Set(1, 256)
	if err != bitmapindex.ErrBitWidthExceeded {
		t.Fatalf("expected ErrBitWidthExceeded for 256 in 8-bit index, got: %v", err)
	}
}

func TestBSI_Concurrency(t *testing.T) {
	idx, _ := bitmapindex.New(16)
	var wg sync.WaitGroup

	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := uint32(0); i < 100; i++ {
				rowID := uint32(workerID*1000) + i
				_ = idx.Set(rowID, uint64(i*10))
				_, _ = idx.Get(rowID)
				_ = idx.Sum(nil)
			}
		}(w)
	}

	wg.Wait()
}

