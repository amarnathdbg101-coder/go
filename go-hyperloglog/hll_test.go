package hyperloglog_test

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-hyperloglog"
)

func TestHLL_SmallCardinality(t *testing.T) {
	sketch, err := hyperloglog.New(hyperloglog.WithPrecision(14))
	if err != nil {
		t.Fatalf("failed to create sketch: %v", err)
	}

	if est := sketch.Estimate(); est != 0 {
		t.Fatalf("expected 0 for empty sketch, got %d", est)
	}

	for i := 0; i < 50; i++ {
		sketch.AddString(fmt.Sprintf("user_%d", i))
	}

	est := sketch.Estimate()
	// For 50 items with LinearCounting, error should be very small (< 5 items)
	diff := math.Abs(float64(est) - 50.0)
	if diff > 5 {
		t.Fatalf("expected estimate ~50, got %d (diff: %f)", est, diff)
	}
}

func TestHLL_MediumCardinality(t *testing.T) {
	sketch, err := hyperloglog.New(hyperloglog.WithPrecision(14))
	if err != nil {
		t.Fatal(err)
	}

	const total = 10000
	for i := 0; i < total; i++ {
		sketch.AddString(fmt.Sprintf("stream_item_%d", i))
	}

	est := sketch.Estimate()
	// Precision 14 has standard error ~ 1.04 / sqrt(16384) = 0.81%
	// 3 * standard error = 2.43% tolerance
	tolerance := float64(total) * 0.03
	diff := math.Abs(float64(est) - float64(total))

	if diff > tolerance {
		t.Fatalf("estimate %d exceeded 3%% tolerance (%f) from %d", est, tolerance, total)
	}
}

func TestHLL_Idempotency(t *testing.T) {
	sketch, _ := hyperloglog.New(hyperloglog.WithPrecision(14))

	// Insert the exact same 10 items 1000 times
	for loop := 0; loop < 1000; loop++ {
		for i := 0; i < 10; i++ {
			sketch.AddString(fmt.Sprintf("repeat_item_%d", i))
		}
	}

	est := sketch.Estimate()
	if est < 9 || est > 11 {
		t.Fatalf("idempotency check failed: expected ~10, got %d", est)
	}
}

func TestHLL_Merge(t *testing.T) {
	sketchA, _ := hyperloglog.New(hyperloglog.WithPrecision(14))
	sketchB, _ := hyperloglog.New(hyperloglog.WithPrecision(14))

	// A: 0..4999 (5000 items)
	for i := 0; i < 5000; i++ {
		sketchA.AddString(fmt.Sprintf("item_%d", i))
	}

	// B: 3000..7999 (5000 items, 2000 overlap)
	for i := 3000; i < 8000; i++ {
		sketchB.AddString(fmt.Sprintf("item_%d", i))
	}

	// Union should have 8000 distinct items
	err := sketchA.Merge(sketchB)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	est := sketchA.Estimate()
	diff := math.Abs(float64(est) - 8000.0)
	tolerance := 8000.0 * 0.03

	if diff > tolerance {
		t.Fatalf("merged estimate %d outside tolerance of 8000 (diff: %f)", est, diff)
	}
}

func TestHLL_Concurrency(t *testing.T) {
	sketch, _ := hyperloglog.New(hyperloglog.WithPrecision(12))
	var wg sync.WaitGroup

	workers := 10
	itemsPerWorker := 500

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < itemsPerWorker; i++ {
				sketch.AddString(fmt.Sprintf("worker_%d_item_%d", workerID, i))
			}
		}(w)
	}

	wg.Wait()

	totalExpected := float64(workers * itemsPerWorker)
	est := sketch.Estimate()
	diff := math.Abs(float64(est) - totalExpected)
	tolerance := totalExpected * 0.05

	if diff > tolerance {
		t.Fatalf("concurrent estimate %d outside tolerance (diff: %f)", est, diff)
	}
}

