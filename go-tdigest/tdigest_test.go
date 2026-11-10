package tdigest_test

import (
	"math"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-tdigest"
)

func TestTDigest_UniformQuantiles(t *testing.T) {
	td, err := tdigest.New(tdigest.WithCompression(200))
	if err != nil {
		t.Fatalf("failed to create t-digest: %v", err)
	}

	const n = 10000
	for i := 1; i <= n; i++ {
		td.Add(float64(i))
	}

	if td.Count() != n {
		t.Fatalf("expected count %d, got %f", n, td.Count())
	}
	if td.Min() != 1.0 || td.Max() != float64(n) {
		t.Fatalf("min/max mismatch: min=%f, max=%f", td.Min(), td.Max())
	}

	// P50 should be ~ 5000 (within 1% tolerance)
	p50, err := td.Quantile(0.5)
	if err != nil || math.Abs(p50-5000.0) > 100 {
		t.Fatalf("P50 expected ~5000, got %f (err: %v)", p50, err)
	}

	// P90 should be ~ 9000
	p90, err := td.Quantile(0.9)
	if err != nil || math.Abs(p90-9000.0) > 100 {
		t.Fatalf("P90 expected ~9000, got %f (err: %v)", p90, err)
	}

	// P99 tail should be ~ 9900 (within 0.5% tolerance due to tail accuracy)
	p99, err := td.Quantile(0.99)
	if err != nil || math.Abs(p99-9900.0) > 50 {
		t.Fatalf("P99 expected ~9900, got %f (err: %v)", p99, err)
	}
}

func TestTDigest_Merge(t *testing.T) {
	td1, _ := tdigest.New(tdigest.WithCompression(100))
	td2, _ := tdigest.New(tdigest.WithCompression(100))

	// td1: 1..5000
	for i := 1; i <= 5000; i++ {
		td1.Add(float64(i))
	}

	// td2: 5001..10000
	for i := 5001; i <= 10000; i++ {
		td2.Add(float64(i))
	}

	// Merge td2 into td1
	if err := td1.Merge(td2); err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	if td1.Count() != 10000 {
		t.Fatalf("expected 10000 count after merge, got %f", td1.Count())
	}

	// P50 of merged 1..10000 should be ~ 5000
	p50, err := td1.Quantile(0.5)
	if err != nil || math.Abs(p50-5000.0) > 150 {
		t.Fatalf("merged P50 expected ~5000, got %f", p50)
	}
}

func TestTDigest_Concurrency(t *testing.T) {
	td, _ := tdigest.New(tdigest.WithCompression(100))
	var wg sync.WaitGroup

	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				td.Add(float64(id*1000 + i))
				_, _ = td.Quantile(0.95)
			}
		}(w)
	}

	wg.Wait()
}

