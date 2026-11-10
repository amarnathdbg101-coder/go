package distlock_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-distlock"
)

func TestLockAndRelease(t *testing.T) {
	locker := distlock.NewLocker(distlock.NewMemoryDriver())
	ctx := context.Background()

	lock, err := locker.Acquire(ctx, "resource_1", distlock.WithTTL(1*time.Second))
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	if lock.Key() != "resource_1" {
		t.Errorf("unexpected key: %s", lock.Key())
	}
	if lock.FencingToken() == 0 {
		t.Errorf("fencing token must be non-zero")
	}

	if err := lock.Release(); err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}
}

func TestMutualExclusion(t *testing.T) {
	locker := distlock.NewLocker(distlock.NewMemoryDriver())
	ctx := context.Background()

	lock1, err := locker.Acquire(ctx, "exclusive_key", distlock.WithTTL(1*time.Second), distlock.WithOwner("worker_A"))
	if err != nil {
		t.Fatalf("worker A failed to acquire lock: %v", err)
	}
	defer lock1.Release()

	// Worker B attempts to acquire same key -> must fail immediately
	_, err2 := locker.Acquire(ctx, "exclusive_key", distlock.WithTTL(1*time.Second), distlock.WithOwner("worker_B"))
	if !errors.Is(err2, distlock.ErrLockNotAcquired) {
		t.Fatalf("expected ErrLockNotAcquired for worker B, got %v", err2)
	}
}

func TestMonotonicFencingTokens(t *testing.T) {
	locker := distlock.NewLocker(distlock.NewMemoryDriver())
	ctx := context.Background()

	lock1, err := locker.Acquire(ctx, "shared_job", distlock.WithTTL(1*time.Second))
	if err != nil {
		t.Fatalf("lock 1 failed: %v", err)
	}
	token1 := lock1.FencingToken()
	_ = lock1.Release()

	lock2, err := locker.Acquire(ctx, "shared_job", distlock.WithTTL(1*time.Second))
	if err != nil {
		t.Fatalf("lock 2 failed: %v", err)
	}
	token2 := lock2.FencingToken()
	_ = lock2.Release()

	if token2 <= token1 {
		t.Fatalf("fencing tokens must be strictly monotonic: token1=%d token2=%d", token1, token2)
	}
}

func TestAutoHeartbeatRenewal(t *testing.T) {
	locker := distlock.NewLocker(distlock.NewMemoryDriver())
	ctx := context.Background()

	// TTL is 25ms, but AutoRenew refreshes it every 8ms
	lock, err := locker.Acquire(ctx, "long_task", distlock.WithTTL(25*time.Millisecond), distlock.WithAutoRenew(true))
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	defer lock.Release()

	// Sleep for 60ms (well past the 25ms initial TTL)
	time.Sleep(60 * time.Millisecond)

	// Another worker should STILL not be able to acquire because heartbeat extended it!
	_, err2 := locker.Acquire(ctx, "long_task", distlock.WithOwner("worker_other"))
	if !errors.Is(err2, distlock.ErrLockNotAcquired) {
		t.Fatalf("expected lock to remain held due to heartbeat renewal, got: %v", err2)
	}
}

func TestRetryAcquire(t *testing.T) {
	locker := distlock.NewLocker(distlock.NewMemoryDriver())
	ctx := context.Background()

	lock1, _ := locker.Acquire(ctx, "retry_key", distlock.WithTTL(1*time.Second), distlock.WithOwner("first"))

	var worker2Acquired int32
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		// Worker 2 polls with a 200ms timeout
		lock2, err := locker.Acquire(ctx, "retry_key",
			distlock.WithOwner("second"),
			distlock.WithRetry(200*time.Millisecond, 10*time.Millisecond),
		)
		if err == nil {
			atomic.AddInt32(&worker2Acquired, 1)
			_ = lock2.Release()
		}
	}()

	// Release lock1 after 20ms
	time.Sleep(20 * time.Millisecond)
	_ = lock1.Release()

	wg.Wait()

	if atomic.LoadInt32(&worker2Acquired) != 1 {
		t.Fatalf("worker 2 should have acquired the lock after release via retry")
	}
}

