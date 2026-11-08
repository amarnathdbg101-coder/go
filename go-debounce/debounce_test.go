package godebounce

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDebouncer_Burst(t *testing.T) {
	d := New(40 * time.Millisecond)
	var count atomic.Int32

	// Fire 10 rapid triggers within 20ms
	for i := 0; i < 10; i++ {
		d.Trigger(func() {
			count.Add(1)
		})
		time.Sleep(2 * time.Millisecond)
	}

	// Should not have fired yet
	if count.Load() != 0 {
		t.Errorf("expected 0 executions before wait window, got %d", count.Load())
	}

	// Wait for debounce period of silence
	time.Sleep(80 * time.Millisecond)

	// Must have fired exactly once
	if count.Load() != 1 {
		t.Errorf("expected exactly 1 execution after debounce, got %d", count.Load())
	}
}

func TestDebouncer_Cancel(t *testing.T) {
	d := New(40 * time.Millisecond)
	var count atomic.Int32

	d.Trigger(func() {
		count.Add(1)
	})

	time.Sleep(10 * time.Millisecond)
	d.Cancel()

	time.Sleep(60 * time.Millisecond)

	if count.Load() != 0 {
		t.Errorf("expected 0 executions after Cancel, got %d", count.Load())
	}
}

func TestThrottler_Window(t *testing.T) {
	th := NewThrottler(40 * time.Millisecond)
	var count atomic.Int32

	// First trigger should execute immediately
	executed := th.Trigger(func() {
		count.Add(1)
	})
	if !executed || count.Load() != 1 {
		t.Fatalf("expected first trigger to execute immediately")
	}

	// Second rapid trigger should be throttled
	executed = th.Trigger(func() {
		count.Add(1)
	})
	if executed || count.Load() != 1 {
		t.Fatalf("expected rapid trigger to be throttled")
	}

	// Wait for throttle window to elapse
	time.Sleep(50 * time.Millisecond)

	// Third trigger should now execute
	executed = th.Trigger(func() {
		count.Add(1)
	})
	if !executed || count.Load() != 2 {
		t.Fatalf("expected trigger after window to execute, count=%d", count.Load())
	}
}

func TestKeyedDebouncer_Isolation(t *testing.T) {
	kd := NewKeyed(40 * time.Millisecond)
	var countA atomic.Int32
	var countB atomic.Int32

	// Rapid triggers for user A
	for i := 0; i < 5; i++ {
		kd.Trigger("user_A", func() {
			countA.Add(1)
		})
	}

	// Rapid triggers for user B
	for i := 0; i < 5; i++ {
		kd.Trigger("user_B", func() {
			countB.Add(1)
		})
	}

	time.Sleep(80 * time.Millisecond)

	if countA.Load() != 1 || countB.Load() != 1 {
		t.Errorf("expected 1 execution per key, got A=%d, B=%d", countA.Load(), countB.Load())
	}
}
