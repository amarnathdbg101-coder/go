package goexponentialbackoff

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackoff_NoJitter(t *testing.T) {
	b := New(Config{
		Base:      100 * time.Millisecond,
		Max:       1000 * time.Millisecond,
		Factor:    2.0,
		Algorithm: NoJitter,
	})

	d0 := b.Next(0)
	if d0 != 100*time.Millisecond {
		t.Errorf("attempt 0: expected 100ms, got %v", d0)
	}

	d1 := b.Next(1)
	if d1 != 200*time.Millisecond {
		t.Errorf("attempt 1: expected 200ms, got %v", d1)
	}

	d2 := b.Next(2)
	if d2 != 400*time.Millisecond {
		t.Errorf("attempt 2: expected 400ms, got %v", d2)
	}

	// Max clamping
	d10 := b.Next(10)
	if d10 != 1000*time.Millisecond {
		t.Errorf("attempt 10: expected clamped 1000ms, got %v", d10)
	}
}

func TestBackoff_FullJitterBounds(t *testing.T) {
	b := New(Config{
		Base:      50 * time.Millisecond,
		Max:       500 * time.Millisecond,
		Factor:    2.0,
		Algorithm: FullJitter,
	})

	for i := 0; i < 20; i++ {
		d := b.Next(2) // ceiling = 200ms
		if d < 0 || d > 200*time.Millisecond {
			t.Errorf("FullJitter outside [0, 200ms]: got %v", d)
		}
	}
}

func TestBackoff_EqualJitterBounds(t *testing.T) {
	b := New(Config{
		Base:      100 * time.Millisecond,
		Max:       1000 * time.Millisecond,
		Factor:    2.0,
		Algorithm: EqualJitter,
	})

	// attempt 1: ceiling = 200ms, half = 100ms, range = [100ms, 200ms]
	for i := 0; i < 20; i++ {
		d := b.Next(1)
		if d < 100*time.Millisecond || d > 200*time.Millisecond {
			t.Errorf("EqualJitter outside [100ms, 200ms]: got %v", d)
		}
	}
}

func TestBackoff_SleepContextCancellation(t *testing.T) {
	b := New(DefaultConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	// Request 5-second backoff sleep, but context expires in 20ms
	err := b.Sleep(ctx, 10)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}
