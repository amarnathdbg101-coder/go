package timingwheel

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTimingWheel_BasicExecution(t *testing.T) {
	// 20ms tick, 10 slots
	tw := New(20*time.Millisecond, 10)
	tw.Start()
	defer tw.Stop()

	done := make(chan struct{})
	start := time.Now()

	_, err := tw.AfterFunc(60*time.Millisecond, func() {
		close(done)
	})
	if err != nil {
		t.Fatalf("failed to schedule timer: %v", err)
	}

	select {
	case <-done:
		elapsed := time.Since(start)
		if elapsed < 40*time.Millisecond {
			t.Errorf("timer fired too early: %v", elapsed)
		}
	case <-time.After(200 * time.Millisecond):
		t.Errorf("timer did not fire within deadline")
	}
}

func TestTimingWheel_Cancel(t *testing.T) {
	tw := New(20*time.Millisecond, 10)
	tw.Start()
	defer tw.Stop()

	var executed int32

	timer, err := tw.AfterFunc(100*time.Millisecond, func() {
		atomic.StoreInt32(&executed, 1)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cancel after 20ms
	time.Sleep(20 * time.Millisecond)
	if !timer.Stop() {
		t.Errorf("expected timer.Stop() to succeed")
	}

	// Wait past the original 100ms deadline
	time.Sleep(120 * time.Millisecond)

	if atomic.LoadInt32(&executed) == 1 {
		t.Errorf("cancelled timer was executed!")
	}
}

func TestTimingWheel_MultiRound(t *testing.T) {
	// Wheel duration = 10ms * 5 = 50ms per revolution.
	// Schedule for 120ms (more than 2 full rounds).
	tw := New(10*time.Millisecond, 5)
	tw.Start()
	defer tw.Stop()

	done := make(chan struct{})
	start := time.Now()

	tw.AfterFunc(120*time.Millisecond, func() {
		close(done)
	})

	select {
	case <-done:
		elapsed := time.Since(start)
		if elapsed < 100*time.Millisecond {
			t.Errorf("multi-round timer fired too early: %v", elapsed)
		}
	case <-time.After(300 * time.Millisecond):
		t.Errorf("multi-round timer did not fire within expected window")
	}
}

func TestTimingWheel_Concurrency(t *testing.T) {
	tw := New(10*time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	const totalTimers = 100
	var count int64
	var wg sync.WaitGroup

	for i := 0; i < totalTimers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tw.AfterFunc(30*time.Millisecond, func() {
				atomic.AddInt64(&count, 1)
			})
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&count) != totalTimers {
		t.Errorf("expected %d timers to execute, got %d", totalTimers, count)
	}
}
