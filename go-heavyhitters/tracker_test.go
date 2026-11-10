package heavyhitters_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-heavyhitters"
)

func TestHeavyHitters_Exact(t *testing.T) {
	tracker, err := heavyhitters.New(heavyhitters.WithCapacity(10))
	if err != nil {
		t.Fatalf("failed to create tracker: %v", err)
	}

	tracker.AddWithCount("product_shoes", 100)
	tracker.AddWithCount("product_shirt", 50)
	tracker.AddWithCount("product_hat", 20)

	top := tracker.TopK(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 items, got %d", len(top))
	}

	if top[0].Key != "product_shoes" || top[0].Count != 100 || top[0].Error != 0 {
		t.Fatalf("unexpected top 1: %+v", top[0])
	}
	if top[1].Key != "product_shirt" || top[1].Count != 50 || top[1].Error != 0 {
		t.Fatalf("unexpected top 2: %+v", top[1])
	}
}

func TestHeavyHitters_StreamingSkewedDistribution(t *testing.T) {
	// Capacity 50 tracking streaming heavy hitters
	tracker, err := heavyhitters.New(heavyhitters.WithCapacity(50))
	if err != nil {
		t.Fatal(err)
	}

	// Dominant heavy hitters
	tracker.AddWithCount("heavy_alpha", 5000)
	tracker.AddWithCount("heavy_beta", 3000)
	tracker.AddWithCount("heavy_gamma", 1500)

	// Stream 10,000 sparse random tail items
	for i := 0; i < 10000; i++ {
		tracker.Add(fmt.Sprintf("tail_noise_%d", i))
	}

	top3 := tracker.TopK(3)
	if len(top3) != 3 {
		t.Fatalf("expected 3 items, got %d", len(top3))
	}

	if top3[0].Key != "heavy_alpha" {
		t.Fatalf("expected heavy_alpha at rank 0, got %s", top3[0].Key)
	}
	if top3[1].Key != "heavy_beta" {
		t.Fatalf("expected heavy_beta at rank 1, got %s", top3[1].Key)
	}
	if top3[2].Key != "heavy_gamma" {
		t.Fatalf("expected heavy_gamma at rank 2, got %s", top3[2].Key)
	}

	// Verify error bound
	for _, item := range top3 {
		if item.Error > item.Count {
			t.Fatalf("error bound violated for %s: count=%d, err=%d", item.Key, item.Count, item.Error)
		}
	}
}

func TestHeavyHitters_Concurrency(t *testing.T) {
	tracker, _ := heavyhitters.New(heavyhitters.WithCapacity(100))
	var wg sync.WaitGroup

	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				tracker.Add("common_heavy_item")
				tracker.Add(fmt.Sprintf("worker_%d_item_%d", id, i))
			}
		}(w)
	}

	wg.Wait()

	top := tracker.TopK(1)
	if len(top) != 1 || top[0].Key != "common_heavy_item" {
		t.Fatalf("expected common_heavy_item at top, got: %+v", top)
	}
}

