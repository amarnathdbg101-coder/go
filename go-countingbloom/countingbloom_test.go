package countingbloom

import (
	"sync"
	"testing"
)

func TestCountingBloom_Basic(t *testing.T) {
	filter, err := New(1000, 0.01)
	if err != nil {
		t.Fatalf("failed to create counting bloom filter: %v", err)
	}

	filter.AddString("user:101")
	filter.AddString("user:102")

	if !filter.ContainsString("user:101") {
		t.Fatalf("expected user:101 to be present")
	}
	if !filter.ContainsString("user:102") {
		t.Fatalf("expected user:102 to be present")
	}
	if filter.ContainsString("user:999") {
		t.Fatalf("did not expect user:999 to be present")
	}

	// Remove user:101
	removed := filter.RemoveString("user:101")
	if !removed {
		t.Fatalf("expected RemoveString to return true")
	}

	// Verify user:101 is gone
	if filter.ContainsString("user:101") {
		t.Fatalf("expected user:101 to be removed")
	}

	// Verify user:102 is STILL present (no false corruption)
	if !filter.ContainsString("user:102") {
		t.Fatalf("user:102 should still be present after user:101 deletion")
	}
}

func TestCountingBloom_RemoveNonExistent(t *testing.T) {
	filter, _ := New(500, 0.01)
	removed := filter.RemoveString("ghost")
	if removed {
		t.Fatalf("expected RemoveString to return false on missing item")
	}
}

func TestCountingBloom_Saturation(t *testing.T) {
	filter, _ := New(100, 0.05)

	// Add same item 300 times (exceeding 8-bit max 255)
	for i := 0; i < 300; i++ {
		filter.AddString("heavy_item")
	}

	if !filter.ContainsString("heavy_item") {
		t.Fatalf("expected heavy_item to be present")
	}
}

func TestCountingBloom_Clear(t *testing.T) {
	filter, _ := New(100, 0.01)
	filter.AddString("alpha")
	filter.AddString("beta")

	filter.Clear()
	if filter.Count() != 0 {
		t.Fatalf("expected count 0 after clear")
	}
	if filter.ContainsString("alpha") || filter.ContainsString("beta") {
		t.Fatalf("expected all items gone after clear")
	}
}

func TestCountingBloom_Concurrent(t *testing.T) {
	filter, _ := NewConcurrent(1000, 0.01)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "item"
			filter.AddString(key)
			filter.ContainsString(key)
			filter.RemoveString(key)
		}(i)
	}
	wg.Wait()
}
