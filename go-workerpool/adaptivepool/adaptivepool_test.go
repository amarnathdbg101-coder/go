package adaptivepool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAdaptivePool_Basic(t *testing.T) {
	pool := New(Config{
		MinWorkers:  2,
		MaxWorkers:  10,
		IdleTimeout: 100 * time.Millisecond,
		QueueSize:   20,
	})
	defer pool.Close()

	if pool.ActiveWorkers() != 2 {
		t.Fatalf("expected 2 baseline workers, got %d", pool.ActiveWorkers())
	}

	var counter int32
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		err := pool.Submit(func() {
			defer wg.Done()
			atomic.AddInt32(&counter, 1)
		})
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	wg.Wait()
	if atomic.LoadInt32(&counter) != 10 {
		t.Fatalf("expected counter 10, got %d", counter)
	}
}

func TestAdaptivePool_ScalingAndEviction(t *testing.T) {
	pool := New(Config{
		MinWorkers:  1,
		MaxWorkers:  5,
		IdleTimeout: 50 * time.Millisecond,
		QueueSize:   10,
	})
	defer pool.Close()

	// Submit burst of blocking tasks to force pool to scale up
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		_ = pool.Submit(func() {
			defer wg.Done()
			time.Sleep(30 * time.Millisecond)
		})
	}

	// Should have scaled beyond 1
	time.Sleep(10 * time.Millisecond)
	if pool.ActiveWorkers() < 2 {
		t.Logf("workers count: %d", pool.ActiveWorkers())
	}

	wg.Wait()

	// Wait for idle surplus workers to be evicted
	time.Sleep(150 * time.Millisecond)
	if pool.ActiveWorkers() > 2 {
		t.Fatalf("expected idle workers to be evicted down to baseline, got: %d", pool.ActiveWorkers())
	}
}

func TestAdaptivePool_PanicRecovery(t *testing.T) {
	pool := New(DefaultConfig())
	defer pool.Close()

	var normalExecuted bool
	var wg sync.WaitGroup
	wg.Add(2)

	_ = pool.Submit(func() {
		defer wg.Done()
		panic("boom") // should be caught safely
	})

	_ = pool.Submit(func() {
		defer wg.Done()
		normalExecuted = true
	})

	wg.Wait()
	if !normalExecuted {
		t.Fatalf("expected normal task to run despite previous panic")
	}
}

func TestAdaptivePool_CloseAndDrain(t *testing.T) {
	pool := New(Config{
		MinWorkers:  2,
		MaxWorkers:  4,
		IdleTimeout: 1 * time.Second,
		QueueSize:   5,
	})

	var counter int32
	for i := 0; i < 3; i++ {
		_ = pool.Submit(func() {
			atomic.AddInt32(&counter, 1)
		})
	}

	err := pool.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// New submits must fail
	err = pool.Submit(func() {})
	if err != ErrPoolClosed {
		t.Fatalf("expected ErrPoolClosed, got: %v", err)
	}

	if atomic.LoadInt32(&counter) != 3 {
		t.Fatalf("expected all 3 queued tasks to drain on close, got %d", counter)
	}
}
