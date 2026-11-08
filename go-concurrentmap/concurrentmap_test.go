package goconcurrentmap

import (
	"sync"
	"testing"
)

func TestConcurrentMap_BasicOperations(t *testing.T) {
	cm := New[string, int]()

	cm.Set("alpha", 10)
	cm.Set("beta", 20)

	val, ok := cm.Get("alpha")
	if !ok || val != 10 {
		t.Fatalf("expected 10 for alpha, got %d (ok=%v)", val, ok)
	}

	if !cm.Has("beta") {
		t.Errorf("expected Has(beta) = true")
	}

	if cm.Has("gamma") {
		t.Errorf("expected Has(gamma) = false")
	}

	if cm.Count() != 2 {
		t.Errorf("expected count 2, got %d", cm.Count())
	}

	// Pop
	popped, ok := cm.Pop("alpha")
	if !ok || popped != 10 {
		t.Errorf("expected pop 10, got %d", popped)
	}
	if cm.Has("alpha") {
		t.Errorf("alpha should not exist after pop")
	}

	// Remove
	removed := cm.Remove("beta")
	if !removed {
		t.Errorf("expected remove to return true")
	}
	if !cm.IsEmpty() {
		t.Errorf("expected map to be empty")
	}
}

func TestConcurrentMap_GetOrSet(t *testing.T) {
	cm := New[string, string]()

	v1, loaded := cm.GetOrSet("key", "val1")
	if loaded || v1 != "val1" {
		t.Errorf("expected loaded=false, val='val1', got loaded=%v, val=%s", loaded, v1)
	}

	v2, loaded := cm.GetOrSet("key", "val2")
	if !loaded || v2 != "val1" {
		t.Errorf("expected loaded=true, val='val1', got loaded=%v, val=%s", loaded, v2)
	}
}

func TestConcurrentMap_Upsert(t *testing.T) {
	cm := New[string, int]()

	// First insert
	val := cm.Upsert("counter", 1, func(exists bool, oldVal, newVal int) int {
		if !exists {
			return newVal
		}
		return oldVal + newVal
	})
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	// Second update (increment)
	val = cm.Upsert("counter", 5, func(exists bool, oldVal, newVal int) int {
		return oldVal + newVal
	})
	if val != 6 {
		t.Errorf("expected 6, got %d", val)
	}
}

func TestConcurrentMap_Iter(t *testing.T) {
	cm := New[int, int]()
	for i := 0; i < 10; i++ {
		cm.Set(i, i*10)
	}

	visited := 0
	cm.Iter(func(key int, val int) bool {
		visited++
		if visited == 5 {
			return false // Early stop
		}
		return true
	})

	if visited != 5 {
		t.Errorf("expected early stop at 5, visited=%d", visited)
	}
}

func TestConcurrentMap_ConcurrentStress(t *testing.T) {
	cm := New[int, int]()
	var wg sync.WaitGroup
	workers := 20
	iterations := 200

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := j
				cm.Upsert(key, 1, func(exists bool, oldVal, newVal int) int {
					if !exists {
						return 1
					}
					return oldVal + 1
				})
			}
		}(i)
	}

	wg.Wait()

	if cm.Count() != iterations {
		t.Errorf("expected %d unique keys, got %d", iterations, cm.Count())
	}

	// Check key 0 has count = workers
	val, ok := cm.Get(0)
	if !ok || val != workers {
		t.Errorf("expected key 0 value = %d, got %d", workers, val)
	}
}

func BenchmarkConcurrentMap_WriteRead(b *testing.B) {
	cm := New[int, int]()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cm.Set(i%1000, i)
			_, _ = cm.Get(i % 1000)
			i++
		}
	})
}
