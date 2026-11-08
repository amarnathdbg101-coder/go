package lfu

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLFU_Eviction(t *testing.T) {
	// Capacity 3
	c := New[string, int](3)

	c.Set("a", 1) // freq 1
	c.Set("b", 2) // freq 1
	c.Set("c", 3) // freq 1

	// Access "a" 3 times -> freq 4
	c.Get("a")
	c.Get("a")
	c.Get("a")

	// Access "b" 1 time -> freq 2
	c.Get("b")

	// "c" remains freq 1.
	// Adding "d" should evict "c"!
	c.Set("d", 4)

	if c.Len() != 3 {
		t.Errorf("expected len 3, got %d", c.Len())
	}
	if _, ok := c.Get("c"); ok {
		t.Errorf("expected 'c' to be evicted as least frequently used")
	}
	if val, ok := c.Get("a"); !ok || val != 1 {
		t.Errorf("expected 'a' to remain in cache")
	}
	if val, ok := c.Get("b"); !ok || val != 2 {
		t.Errorf("expected 'b' to remain in cache")
	}
	if val, ok := c.Get("d"); !ok || val != 4 {
		t.Errorf("expected 'd' to be present")
	}
}

func TestLFU_TieBreaking(t *testing.T) {
	// Capacity 2
	c := New[string, int](2)

	c.Set("x", 10) // freq 1 (older)
	c.Set("y", 20) // freq 1 (newer)

	// Both have freq 1.
	// Adding "z" should evict the older one ("x")!
	c.Set("z", 30)

	if _, ok := c.Get("x"); ok {
		t.Errorf("expected 'x' to be evicted by tie-breaker")
	}
	if _, ok := c.Get("y"); !ok {
		t.Errorf("expected 'y' to remain")
	}
	if _, ok := c.Get("z"); !ok {
		t.Errorf("expected 'z' to remain")
	}
}

func TestLFU_TTL(t *testing.T) {
	c := New[string, string](5)

	c.SetWithTTL("temp", "value", 40*time.Millisecond)
	c.Set("perm", "value")

	val, ok := c.Get("temp")
	if !ok || val != "value" {
		t.Fatalf("expected temp to be valid initially")
	}

	time.Sleep(60 * time.Millisecond)

	// temp should now be expired
	if _, ok := c.Get("temp"); ok {
		t.Errorf("expected temp to have expired")
	}
	if _, ok := c.Get("perm"); !ok {
		t.Errorf("expected perm to still be valid")
	}
}

func TestLFU_PurgeExpired(t *testing.T) {
	c := New[string, int](5)

	c.SetWithTTL("k1", 1, 30*time.Millisecond)
	c.SetWithTTL("k2", 2, 30*time.Millisecond)
	c.Set("k3", 3)

	time.Sleep(50 * time.Millisecond)

	evicted := c.PurgeExpired()
	if evicted != 2 {
		t.Errorf("expected 2 evicted keys, got %d", evicted)
	}
	if c.Len() != 1 {
		t.Errorf("expected 1 remaining key, got %d", c.Len())
	}
}

func TestLFU_DeleteAndClear(t *testing.T) {
	c := New[int, int](3)
	c.Set(1, 100)
	c.Set(2, 200)

	if !c.Delete(1) {
		t.Errorf("expected delete to return true")
	}
	if c.Delete(999) {
		t.Errorf("expected delete nonexistent to return false")
	}
	if c.Len() != 1 {
		t.Errorf("expected len 1, got %d", c.Len())
	}

	c.Clear()
	if c.Len() != 0 {
		t.Errorf("expected len 0 after clear")
	}
}

func TestLFU_Concurrency(t *testing.T) {
	c := New[string, int](20)
	var wg sync.WaitGroup
	const routines = 40
	const iterations = 50

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key-%d", j%30)
				c.Set(key, j)
				c.Get(key)
			}
		}(i)
	}

	wg.Wait()

	if c.Len() > c.Capacity() {
		t.Errorf("cache size exceeded capacity: len=%d, cap=%d", c.Len(), c.Capacity())
	}
}
