package latch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCountDownLatch_Basic(t *testing.T) {
	cdl := NewCountDownLatch(3)
	if cdl.GetCount() != 3 {
		t.Errorf("expected count 3, got %d", cdl.GetCount())
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		cdl.CountDown()
		cdl.CountDown()
		cdl.CountDown()
	}()

	cdl.Wait()
	if cdl.GetCount() != 0 {
		t.Errorf("expected count 0 after wait, got %d", cdl.GetCount())
	}
}

func TestCountDownLatch_Timeout(t *testing.T) {
	cdl := NewCountDownLatch(1)

	// Should timeout
	if cdl.WaitTimeout(30 * time.Millisecond) {
		t.Errorf("expected timeout (false)")
	}

	// Count down and check again
	cdl.CountDown()
	if !cdl.WaitTimeout(10 * time.Millisecond) {
		t.Errorf("expected success after countdown (true)")
	}
}

func TestCountDownLatch_Context(t *testing.T) {
	cdl := NewCountDownLatch(1)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := cdl.WaitContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestCyclicBarrier_MultiCycle(t *testing.T) {
	const parties = 4
	const cycles = 3

	var actionCount int64
	barrier := NewCyclicBarrier(parties, func() {
		atomic.AddInt64(&actionCount, 1)
	})

	var wg sync.WaitGroup

	for i := 0; i < parties; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for c := 0; c < cycles; c++ {
				idx, err := barrier.Await(context.Background())
				if err != nil {
					t.Errorf("worker %d cycle %d await failed: %v", workerID, c, err)
					return
				}
				if idx < 0 || idx >= parties {
					t.Errorf("invalid arrival index: %d", idx)
				}
				// Small jitter before next cycle
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	if actionCount != int64(cycles) {
		t.Errorf("expected barrier action to run %d times, got %d", cycles, actionCount)
	}
}

func TestCyclicBarrier_BrokenOnCancel(t *testing.T) {
	barrier := NewCyclicBarrier(3, nil)

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 2)

	// Party 1 waits with cancellable context
	go func() {
		_, err := barrier.Await(ctx)
		errChan <- err
	}()

	// Party 2 waits with normal context
	go func() {
		time.Sleep(10 * time.Millisecond)
		_, err := barrier.Await(context.Background())
		errChan <- err
	}()

	time.Sleep(30 * time.Millisecond)
	cancel() // Cancelling Party 1 should break the barrier and release Party 2

	err1 := <-errChan
	err2 := <-errChan

	// One should be context.Canceled, the other should be ErrBrokenBarrier
	if !errors.Is(err1, context.Canceled) && !errors.Is(err2, context.Canceled) {
		t.Errorf("expected one error to be context.Canceled, got err1=%v, err2=%v", err1, err2)
	}
	if !errors.Is(err1, ErrBrokenBarrier) && !errors.Is(err2, ErrBrokenBarrier) {
		t.Errorf("expected one error to be ErrBrokenBarrier, got err1=%v, err2=%v", err1, err2)
	}
}
