package slidingstats

import (
	"math"
	"sync"
	"testing"
	"time"
)

func TestSlidingStats_Basic(t *testing.T) {
	w, err := New(1*time.Second, 10)
	if err != nil {
		t.Fatalf("failed to create window: %v", err)
	}

	// Record values 1 to 100
	for i := 1; i <= 100; i++ {
		w.Observe(float64(i))
	}

	snap := w.Snapshot()

	if snap.Count != 100 {
		t.Fatalf("expected count 100, got %d", snap.Count)
	}
	if snap.Min != 1.0 {
		t.Fatalf("expected min 1.0, got %f", snap.Min)
	}
	if snap.Max != 100.0 {
		t.Fatalf("expected max 100.0, got %f", snap.Max)
	}
	if math.Abs(snap.Mean-50.5) > 0.1 {
		t.Fatalf("expected mean ~50.5, got %f", snap.Mean)
	}
	// P50 should be ~50.5
	if math.Abs(snap.P50-50.5) > 1.0 {
		t.Fatalf("expected P50 ~50.5, got %f", snap.P50)
	}
	// P99 should be ~99
	if snap.P99 < 98.0 || snap.P99 > 100.0 {
		t.Fatalf("expected P99 ~99, got %f", snap.P99)
	}
}

func TestSlidingStats_Expiration(t *testing.T) {
	// 100ms window, 5 buckets (20ms each)
	w, _ := New(100*time.Millisecond, 5)

	w.Observe(42.0)
	snap := w.Snapshot()
	if snap.Count != 1 {
		t.Fatalf("expected count 1 before sleep")
	}

	// Sleep past window
	time.Sleep(150 * time.Millisecond)

	snapExpired := w.Snapshot()
	if snapExpired.Count != 0 {
		t.Fatalf("expected count 0 after expiration, got %d", snapExpired.Count)
	}
}

func TestSlidingStats_Empty(t *testing.T) {
	w, _ := New(1*time.Second, 5)
	snap := w.Snapshot()
	if snap.Count != 0 || snap.Mean != 0 {
		t.Fatalf("expected zeroed snapshot for empty window")
	}
}

func TestSlidingStats_Concurrent(t *testing.T) {
	w, _ := New(1*time.Second, 10)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			w.Observe(float64(id * 2))
			_ = w.Snapshot()
		}(i)
	}
	wg.Wait()

	snap := w.Snapshot()
	if snap.Count != 50 {
		t.Fatalf("expected count 50, got %d", snap.Count)
	}
}
