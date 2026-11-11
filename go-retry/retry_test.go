package goretry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetry_SuccessAfterRetries(t *testing.T) {
	var attempts int32
	var onRetryCalls int32

	err := Do(
		context.Background(),
		func(ctx context.Context) error {
			count := atomic.AddInt32(&attempts, 1)
			if count < 3 {
				return errors.New("temporary error")
			}
			return nil
		},
		WithMaxAttempts(5),
		WithBackoff(Constant(5*time.Millisecond)),
		WithJitter(NoJitter),
		WithOnRetry(func(attempt int, err error, delay time.Duration) {
			atomic.AddInt32(&onRetryCalls, 1)
		}),
	)

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
	if atomic.LoadInt32(&onRetryCalls) != 2 {
		t.Fatalf("expected 2 onRetry calls, got %d", atomic.LoadInt32(&onRetryCalls))
	}
}

func TestRetry_Exhaustion(t *testing.T) {
	var attempts int32
	baseErr := errors.New("permanent failure")

	err := Do(
		context.Background(),
		func(ctx context.Context) error {
			atomic.AddInt32(&attempts, 1)
			return baseErr
		},
		WithMaxAttempts(3),
		WithBackoff(Constant(1*time.Millisecond)),
		WithJitter(NoJitter),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var retryErr *RetryError
	if !errors.As(err, &retryErr) {
		t.Fatalf("expected *RetryError, got %v", err)
	}
	if retryErr.Attempts != 3 {
		t.Fatalf("expected 3 attempts recorded in error, got %d", retryErr.Attempts)
	}
	if !errors.Is(err, baseErr) {
		t.Fatal("error unwrap should match baseErr")
	}
}

func TestRetry_RetryIf_StopsEarly(t *testing.T) {
	var attempts int32
	errValidation := errors.New("bad input")

	err := Do(
		context.Background(),
		func(ctx context.Context) error {
			atomic.AddInt32(&attempts, 1)
			return errValidation
		},
		WithMaxAttempts(5),
		WithRetryIf(func(err error) bool {
			// Do NOT retry validation errors
			return !errors.Is(err, errValidation)
		}),
	)

	if !errors.Is(err, errValidation) {
		t.Fatalf("expected errValidation, got %v", err)
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Fatalf("expected exactly 1 attempt because error was non-retriable, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := Do(
		ctx,
		func(ctx context.Context) error {
			return errors.New("timeout trigger")
		},
		WithMaxAttempts(10),
		WithBackoff(Constant(50*time.Millisecond)), // Will exceed context timeout
		WithJitter(NoJitter),
	)

	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context error, got %v", err)
	}
}

func TestDoWithData(t *testing.T) {
	var attempts int32

	res, err := DoWithData(
		context.Background(),
		func(ctx context.Context) (string, error) {
			count := atomic.AddInt32(&attempts, 1)
			if count < 2 {
				return "", errors.New("blip")
			}
			return "SUCCESS_DATA", nil
		},
		WithMaxAttempts(3),
		WithBackoff(Constant(1*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("DoWithData failed: %v", err)
	}
	if res != "SUCCESS_DATA" {
		t.Fatalf("expected SUCCESS_DATA, got %s", res)
	}
}

func TestBackoff_StrategiesAndSleep(t *testing.T) {
	// Exponential
	exp := Exponential(10*time.Millisecond, 100*time.Millisecond, 2.0)
	d0 := exp.NextDelay(0)
	d1 := exp.NextDelay(1)
	if d1 < d0 {
		t.Errorf("expected d1 >= d0, got d0=%v, d1=%v", d0, d1)
	}

	// Decorrelated
	decor := Decorrelated(10*time.Millisecond, 100*time.Millisecond)
	dec0 := decor.NextDelay(0)
	if dec0 < 10*time.Millisecond || dec0 > 100*time.Millisecond {
		t.Errorf("decorrelated out of bounds: %v", dec0)
	}

	// Linear
	lin := Linear(5*time.Millisecond, 5*time.Millisecond, 50*time.Millisecond)
	if lin.NextDelay(2) != 15*time.Millisecond {
		t.Errorf("expected 15ms, got %v", lin.NextDelay(2))
	}

	// Backoff.Sleep with quick success
	ctx := context.Background()
	if err := exp.Sleep(ctx, 0); err != nil {
		t.Errorf("expected nil error on Sleep: %v", err)
	}

	// Backoff.Sleep with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := exp.Sleep(canceledCtx, 1); !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Jitter coverage
	jitFull := ApplyJitter(100*time.Millisecond, FullJitter)
	if jitFull > 100*time.Millisecond {
		t.Errorf("full jitter exceeded base: %v", jitFull)
	}

	jitEqual := ApplyJitter(100*time.Millisecond, EqualJitter)
	if jitEqual < 50*time.Millisecond || jitEqual > 100*time.Millisecond {
		t.Errorf("equal jitter out of bounds: %v", jitEqual)
	}

	jitDec := ApplyJitter(100*time.Millisecond, DecorrelatedJitter)
	if jitDec < 50*time.Millisecond || jitDec > 150*time.Millisecond {
		t.Errorf("decorrelated jitter out of bounds: %v", jitDec)
	}
}

