package syncmap

import (
	"sync"
	"testing"
)

func TestMap_BasicOps(t *testing.T) {
	m := New[string, int]()

	m.Store("apple", 10)
	m.Store("banana", 20)

	if !m.Has("apple") {
		t.Errorf("expected apple to exist")
	}

	val, ok := m.Load("apple")
	if !ok || val != 10 {
		t.Errorf("expected 10, got %d (ok=%v)", val, ok)
	}

	if m.Len() != 2 {
		t.Errorf("expected len 2, got %d", m.Len())
	}

	m.Delete("banana")
	if m.Has("banana") || m.Len() != 1 {
		t.Errorf("expected banana to be deleted")
	}

	m.Clear()
	if m.Len() != 0 {
		t.Errorf("expected len 0 after clear")
	}
}

func TestMap_LoadOrStoreAndSwap(t *testing.T) {
	m := New[int, string]()

	// First time: stores
	val, loaded := m.LoadOrStore(1, "one")
	if loaded || val != "one" {
		t.Errorf("expected loaded=false, val=one, got loaded=%v, val=%s", loaded, val)
	}

	// Second time: loads
	val, loaded = m.LoadOrStore(1, "ONE_NEW")
	if !loaded || val != "one" {
		t.Errorf("expected loaded=true, val=one, got loaded=%v, val=%s", loaded, val)
	}

	// Swap
	prev, hadPrev := m.Swap(1, "one_updated")
	if !hadPrev || prev != "one" || !m.Has(1) {
		t.Errorf("swap mismatch")
	}

	// LoadAndDelete
	delVal, wasLoaded := m.LoadAndDelete(1)
	if !wasLoaded || delVal != "one_updated" || m.Has(1) {
		t.Errorf("load and delete mismatch")
	}
}

func TestMap_CompareOperations(t *testing.T) {
	m := New[string, int]()
	m.Store("counter", 100)

	// CAS
	if !m.CompareAndSwap("counter", 100, 101) {
		t.Errorf("expected CAS to succeed")
	}
	if m.CompareAndSwap("counter", 100, 200) {
		t.Errorf("expected CAS to fail with wrong old value")
	}

	// CAD
	if !m.CompareAndDelete("counter", 101) {
		t.Errorf("expected CAD to succeed")
	}
	if m.Has("counter") {
		t.Errorf("expected counter to be deleted after CAD")
	}
}

func TestMap_KeysValuesRange(t *testing.T) {
	m := New[string, int]()
	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	keys := m.Keys()
	vals := m.Values()

	if len(keys) != 3 || len(vals) != 3 {
		t.Errorf("keys/vals length mismatch: %d, %d", len(keys), len(vals))
	}

	sum := 0
	m.Range(func(k string, v int) bool {
		sum += v
		return true
	})
	if sum != 6 {
		t.Errorf("expected sum 6, got %d", sum)
	}
}

func TestMap_Compute(t *testing.T) {
	m := New[string, int]()

	// Compute new value on empty key
	res := m.Compute("views", func(oldVal int, exists bool) (int, bool) {
		if !exists {
			return 1, false
		}
		return oldVal + 1, false
	})
	if res != 1 {
		t.Errorf("expected 1, got %d", res)
	}

	// Increment existing
	res = m.Compute("views", func(oldVal int, exists bool) (int, bool) {
		return oldVal + 5, false
	})
	if res != 6 {
		t.Errorf("expected 6, got %d", res)
	}

	// Delete via compute
	m.Compute("views", func(oldVal int, exists bool) (int, bool) {
		return 0, true // delete
	})
	if m.Has("views") {
		t.Errorf("expected views to be deleted via compute")
	}
}

func TestMap_Concurrency(t *testing.T) {
	m := New[int, int]()
	var wg sync.WaitGroup
	const workers = 40
	const iterations = 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := j % 10
				m.Compute(key, func(oldVal int, exists bool) (int, bool) {
					return oldVal + 1, false
				})
			}
		}(i)
	}

	wg.Wait()

	totalOperations := 0
	m.Range(func(k, v int) bool {
		totalOperations += v
		return true
	})

	expectedTotal := workers * iterations
	if totalOperations != expectedTotal {
		t.Errorf("expected total operations %d, got %d", expectedTotal, totalOperations)
	}
}
