package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPolicy_HappyPath(t *testing.T) {
	policy := NewPolicy[string]().
		WithTimeout(1 * time.Second)

	res, err := policy.Execute(context.Background(), func(ctx context.Context) (string, error) {
		return "success", nil
	})

	if err != nil || res != "success" {
		t.Fatalf("expected success, got res=%s, err=%v", res, err)
	}
}

func TestPolicy_RetryTransientSuccess(t *testing.T) {
	policy := NewPolicy[int]().
		WithRetry(RetryConfig{
			MaxAttempts:     3,
			InitialInterval: 10 * time.Millisecond,
			Jitter:          false,
		})

	var attempts int32
	res, err := policy.Execute(context.Background(), func(ctx context.Context) (int, error) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return 0, errors.New("transient error")
		}
		return 42, nil
	})

	if err != nil || res != 42 {
		t.Fatalf("expected 42 after retries, got res=%d, err=%v", res, err)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestPolicy_Timeout(t *testing.T) {
	policy := NewPolicy[string]().
		WithTimeout(25 * time.Millisecond)

	_, err := policy.Execute(context.Background(), func(ctx context.Context) (string, error) {
		select {
		case <-time.After(100 * time.Millisecond):
			return "done", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestPolicy_Bulkhead(t *testing.T) {
	policy := NewPolicy[string]().
		WithBulkhead(2) // max 2 concurrent

	var wg sync.WaitGroup
	var blockedCount int64

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := policy.Execute(context.Background(), func(ctx context.Context) (string, error) {
				time.Sleep(50 * time.Millisecond)
				return "ok", nil
			})
			if errors.Is(err, ErrBulkheadFull) {
				atomic.AddInt64(&blockedCount, 1)
			}
		}()
	}

	wg.Wait()

	if atomic.LoadInt64(&blockedCount) == 0 {
		t.Errorf("expected some executions to be blocked by bulkhead")
	}
}

func TestPolicy_CircuitBreaker(t *testing.T) {
	policy := NewPolicy[string]().
		WithCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 2,
			ResetTimeout:     100 * time.Millisecond,
		})

	// Fail twice
	for i := 0; i < 2; i++ {
		policy.Execute(context.Background(), func(ctx context.Context) (string, error) {
			return "", errors.New("upstream failed")
		})
	}

	// Next call should be rejected immediately by open circuit
	_, err := policy.Execute(context.Background(), func(ctx context.Context) (string, error) {
		return "should not run", nil
	})

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got: %v", err)
	}
}

func TestPolicy_Fallback(t *testing.T) {
	policy := NewPolicy[string]().
		WithRetry(RetryConfig{MaxAttempts: 2, InitialInterval: 10 * time.Millisecond}).
		WithFallback(func(ctx context.Context, err error) (string, error) {
			return "cached-fallback", nil
		})

	res, err := policy.Execute(context.Background(), func(ctx context.Context) (string, error) {
		return "", errors.New("fatal payment error")
	})

	if err != nil {
		t.Fatalf("unexpected error when fallback was registered: %v", err)
	}
	if res != "cached-fallback" {
		t.Errorf("expected 'cached-fallback', got '%s'", res)
	}
}
