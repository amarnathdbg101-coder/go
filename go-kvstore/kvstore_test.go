package gokvstore

import (
	"sync"
	"testing"
	"time"
)

func TestStore_BasicGetSetDelete(t *testing.T) {
	s := New(DefaultConfig())
	defer s.Close()

	s.Set("user:1", "Alice", 0)

	val, ok := s.Get("user:1")
	if !ok || val != "Alice" {
		t.Fatalf("expected 'Alice', got %v (ok=%v)", val, ok)
	}

	if s.Len() != 1 {
		t.Errorf("expected len 1, got %d", s.Len())
	}

	deleted := s.Delete("user:1")
	if !deleted {
		t.Errorf("expected delete to return true")
	}

	_, ok = s.Get("user:1")
	if ok {
		t.Errorf("expected key to be deleted")
	}
}

func TestStore_TTLExpiration(t *testing.T) {
	s := New(Config{
		ShardCount:      4,
		CleanupInterval: 20 * time.Millisecond,
	})
	defer s.Close()

	s.Set("session:temp", "secret-token", 50*time.Millisecond)

	// Immediately accessible
	val, ok := s.Get("session:temp")
	if !ok || val != "secret-token" {
		t.Fatalf("expected secret-token immediately, got %v", val)
	}

	// Wait for TTL expiration
	time.Sleep(120 * time.Millisecond)

	_, ok = s.Get("session:temp")
	if ok {
		t.Errorf("expected key to be expired after 120ms")
	}

	if s.Len() != 0 {
		t.Errorf("expected store len 0 after expiration, got %d", s.Len())
	}
}

func TestStore_GetOrSet(t *testing.T) {
	s := New(DefaultConfig())
	defer s.Close()

	actual, loaded := s.GetOrSet("counter", 100, 0)
	if loaded || actual != 100 {
		t.Errorf("expected loaded=false, actual=100, got loaded=%v, actual=%v", loaded, actual)
	}

	// Second call should return existing
	actual, loaded = s.GetOrSet("counter", 200, 0)
	if !loaded || actual != 100 {
		t.Errorf("expected loaded=true, actual=100, got loaded=%v, actual=%v", loaded, actual)
	}
}

func TestStore_CompareAndSwap(t *testing.T) {
	s := New(DefaultConfig())
	defer s.Close()

	s.Set("status", "pending", 0)

	// Fails with mismatched oldVal
	swapped := s.CompareAndSwap("status", "running", "completed")
	if swapped {
		t.Errorf("expected CAS to fail with wrong oldVal")
	}

	// Succeeds with matched oldVal
	swapped = s.CompareAndSwap("status", "pending", "completed")
	if !swapped {
		t.Errorf("expected CAS to succeed")
	}

	val, _ := s.Get("status")
	if val != "completed" {
		t.Errorf("expected status 'completed', got %v", val)
	}
}

func TestStore_Increment(t *testing.T) {
	s := New(DefaultConfig())
	defer s.Close()

	// New key increments from 0
	val, err := s.Increment("views", 1)
	if err != nil || val != 1 {
		t.Fatalf("expected 1, got %d, err %v", val, err)
	}

	// Subsequent increment
	val, err = s.Increment("views", 5)
	if err != nil || val != 6 {
		t.Fatalf("expected 6, got %d, err %v", val, err)
	}

	// Decrement
	val, err = s.Increment("views", -2)
	if err != nil || val != 4 {
		t.Fatalf("expected 4, got %d, err %v", val, err)
	}
}

func TestStore_ConcurrentStress(t *testing.T) {
	s := New(DefaultConfig())
	defer s.Close()

	var wg sync.WaitGroup
	workers := 20
	iterations := 200

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = s.Increment("global_counter", 1)
			}
		}(i)
	}

	wg.Wait()

	val, ok := s.Get("global_counter")
	if !ok || val.(int64) != int64(workers*iterations) {
		t.Errorf("expected total %d, got %v", workers*iterations, val)
	}
}
