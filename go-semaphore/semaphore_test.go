package semaphore

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemaphore_BasicAcquireRelease(t *testing.T) {
	sem := New(5)

	if sem.Available() != 5 || sem.Capacity() != 5 {
		t.Fatalf("expected 5 available and 5 capacity")
	}

	ctx := context.Background()
	if err := sem.Acquire(ctx, 3); err != nil {
		t.Fatalf("unexpected error acquiring 3: %v", err)
	}

	if sem.Available() != 2 {
		t.Errorf("expected 2 available, got %d", sem.Available())
	}

	if err := sem.Release(3); err != nil {
		t.Fatalf("unexpected error releasing 3: %v", err)
	}

	if sem.Available() != 5 {
		t.Errorf("expected 5 available, got %d", sem.Available())
	}
}

func TestSemaphore_TryAcquire(t *testing.T) {
	sem := New(2)

	if !sem.TryAcquire(2) {
		t.Errorf("expected TryAcquire(2) to succeed")
	}

	if sem.TryAcquire(1) {
		t.Errorf("expected TryAcquire(1) to fail when empty")
	}

	sem.Release(2)
	if !sem.TryAcquire(1) {
		t.Errorf("expected TryAcquire(1) to succeed after release")
	}
}

func TestSemaphore_AcquireTimeout(t *testing.T) {
	sem := New(1)
	sem.TryAcquire(1) // Semaphore now empty

	// Timeout expected
	err := sem.AcquireTimeout(1, 50*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got: %v", err)
	}

	// Release in background before timeout
	go func() {
		time.Sleep(20 * time.Millisecond)
		sem.Release(1)
	}()

	err = sem.AcquireTimeout(1, 100*time.Millisecond)
	if err != nil {
		t.Errorf("expected success acquiring with generous timeout, got: %v", err)
	}
}

func TestSemaphore_ContextCancellation(t *testing.T) {
	sem := New(1)
	sem.TryAcquire(1)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := sem.Acquire(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}

	// Verify that waiter queue was cleaned up
	sem.Release(1)
	if sem.Available() != 1 {
		t.Errorf("expected 1 token available after release, got %d", sem.Available())
	}
}

func TestSemaphore_Drain(t *testing.T) {
	sem := New(4)
	sem.TryAcquire(1)

	drained := sem.Drain()
	if drained != 3 {
		t.Errorf("expected 3 drained, got %d", drained)
	}

	if sem.Available() != 0 {
		t.Errorf("expected 0 available after drain")
	}
}

func TestSemaphore_Close(t *testing.T) {
	sem := New(1)
	sem.TryAcquire(1)

	errChan := make(chan error, 1)
	go func() {
		errChan <- sem.Acquire(context.Background(), 1)
	}()

	time.Sleep(30 * time.Millisecond)
	sem.Close()

	err := <-errChan
	if !errors.Is(err, ErrClosed) {
		t.Errorf("expected ErrClosed on close, got: %v", err)
	}
}

func TestSemaphore_ExcessReleaseAndInvalidWeight(t *testing.T) {
	sem := New(2)

	// Release excess
	err := sem.Release(1)
	if !errors.Is(err, ErrReleaseExcess) {
		t.Errorf("expected ErrReleaseExcess, got: %v", err)
	}

	// Invalid weights
	if sem.Acquire(context.Background(), 0) != ErrInvalidWeight {
		t.Errorf("expected ErrInvalidWeight for weight 0")
	}
	if sem.Acquire(context.Background(), 5) != ErrInvalidWeight {
		t.Errorf("expected ErrInvalidWeight for weight > capacity")
	}
}

func TestSemaphore_Concurrency(t *testing.T) {
	const maxConcurrency = 5
	const totalTasks = 100

	sem := New(maxConcurrency)
	var activeCount int64
	var maxObserved int64
	var wg sync.WaitGroup

	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := sem.Acquire(context.Background(), 1); err != nil {
				t.Errorf("acquire failed: %v", err)
				return
			}

			current := atomic.AddInt64(&activeCount, 1)
			// Track max concurrency observed
			for {
				oldMax := atomic.LoadInt64(&maxObserved)
				if current <= oldMax || atomic.CompareAndSwapInt64(&maxObserved, oldMax, current) {
					break
				}
			}

			time.Sleep(2 * time.Millisecond)

			atomic.AddInt64(&activeCount, -1)
			sem.Release(1)
		}()
	}

	wg.Wait()

	if maxObserved > maxConcurrency {
		t.Errorf("concurrency violation: observed %d, limit was %d", maxObserved, maxConcurrency)
	}
}
