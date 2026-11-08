package gocron

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseCron(t *testing.T) {
	// Every 15 mins: "*/15 * * * *"
	cs, err := ParseCron("*/15 * * * *")
	if err != nil {
		t.Fatalf("ParseCron failed: %v", err)
	}

	if !cs.minutes[0] || !cs.minutes[15] || !cs.minutes[30] || !cs.minutes[45] {
		t.Fatalf("minutes not parsed properly: %+v", cs.minutes)
	}
	if cs.minutes[10] {
		t.Fatal("10 should not be in */15")
	}

	// Midnight: "0 0 * * *"
	midnight, err := ParseCron("0 0 * * *")
	if err != nil {
		t.Fatalf("ParseCron midnight failed: %v", err)
	}
	if !midnight.minutes[0] || !midnight.hours[0] {
		t.Fatal("midnight fields mismatch")
	}
}

func TestScheduler_Every(t *testing.T) {
	sched := New()

	var counter int64

	// Every 30ms
	sched.Every("fast_job", 30*time.Millisecond, func() {
		atomic.AddInt64(&counter, 1)
	})

	_ = sched.Start()

	// Wait 150ms (should tick ~3-5 times)
	time.Sleep(150 * time.Millisecond)
	sched.Stop()

	count := atomic.LoadInt64(&counter)
	if count < 2 {
		t.Fatalf("expected at least 2 ticks, got %d", count)
	}
}

func TestScheduler_PanicIsolation(t *testing.T) {
	sched := New()

	var panicReported atomic.Bool
	var once sync.Once
	done := make(chan struct{})

	sched.SetOnPanic(func(name string, r any) {
		if name == "panicking_job" {
			panicReported.Store(true)
			once.Do(func() {
				close(done)
			})
		}
	})

	sched.Every("panicking_job", 10*time.Millisecond, func() {
		panic("nil pointer in scheduled job")
	})

	var healthyCount int64
	sched.Every("healthy_job", 10*time.Millisecond, func() {
		atomic.AddInt64(&healthyCount, 1)
	})

	_ = sched.Start()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for panic callback")
	}

	time.Sleep(50 * time.Millisecond)
	sched.Stop()

	if !panicReported.Load() {
		t.Fatal("panic was not intercepted")
	}
	if atomic.LoadInt64(&healthyCount) == 0 {
		t.Fatal("healthy job should execute even if another job panics")
	}
}
