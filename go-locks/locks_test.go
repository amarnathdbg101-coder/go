package golocks

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestKeyedMutex_ConcurrentDifferentKeys(t *testing.T) {
	km := NewKeyedMutex()

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		unlock := km.Lock("user_1")
		time.Sleep(50 * time.Millisecond)
		unlock()
	}()

	go func() {
		defer wg.Done()
		unlock := km.Lock("user_2")
		time.Sleep(50 * time.Millisecond)
		unlock()
	}()

	wg.Wait()
	duration := time.Since(start)

	// Since keys are different, both should run concurrently in ~50ms (not serialized 100ms)
	if duration >= 90*time.Millisecond {
		t.Fatalf("different keys should execute concurrently, took: %v", duration)
	}

	// Memory leak check: after unlock, locks map must be completely empty!
	if km.ActiveKeys() != 0 {
		t.Fatalf("expected 0 active keys after unlock, got %d", km.ActiveKeys())
	}
}

func TestKeyedMutex_SerializeSameKey(t *testing.T) {
	km := NewKeyedMutex()

	var counter int
	var wg sync.WaitGroup
	const routines = 10
	const perRoutine = 100

	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perRoutine; j++ {
				_ = km.WithLock("shared_account", func() error {
					counter++
					return nil
				})
			}
		}()
	}

	wg.Wait()

	expected := routines * perRoutine
	if counter != expected {
		t.Fatalf("race condition detected: expected %d, got %d", expected, counter)
	}
	if km.ActiveKeys() != 0 {
		t.Fatalf("expected 0 active keys, got %d", km.ActiveKeys())
	}
}

func TestKeyedMutex_TryLock(t *testing.T) {
	km := NewKeyedMutex()

	unlock1, acquired1 := km.TryLock("exclusive_resource")
	if !acquired1 {
		t.Fatal("expected first TryLock to succeed")
	}

	// Second attempt on same key should fail
	unlock2, acquired2 := km.TryLock("exclusive_resource")
	if acquired2 {
		unlock2()
		t.Fatal("second TryLock should have failed")
	}

	unlock1()

	// Third attempt after unlock should succeed
	unlock3, acquired3 := km.TryLock("exclusive_resource")
	if !acquired3 {
		t.Fatal("third TryLock should have succeeded after unlock")
	}
	unlock3()

	if km.ActiveKeys() != 0 {
		t.Fatalf("expected 0 active keys, got %d", km.ActiveKeys())
	}
}

func TestKeyToInt64(t *testing.T) {
	id1 := KeyToInt64("wallet_balance_user_102")
	id2 := KeyToInt64("wallet_balance_user_102")
	id3 := KeyToInt64("wallet_balance_user_103")

	if id1 != id2 {
		t.Fatal("identical keys must produce identical int64 hashes")
	}
	if id1 == id3 {
		t.Fatal("distinct keys should produce different int64 hashes")
	}

	locker := NewPostgresLockerFromString("order_checkout_ORD99")
	if locker.LockID() == 0 {
		t.Fatal("expected non-zero lock ID")
	}
}

func TestDistributedLocker_AcquireAndRelease(t *testing.T) {
	locker := NewDistributedLocker(NewMemoryDriver())
	ctx := context.Background()

	lock, err := locker.Acquire(ctx, "resource_1", WithDistLockTTL(1*time.Second))
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

	// Double release should return ErrLockReleased
	if err := lock.Release(); !errors.Is(err, ErrLockReleased) {
		t.Fatalf("expected ErrLockReleased, got %v", err)
	}
}

func TestDistributedLocker_MutualExclusion(t *testing.T) {
	locker := NewDistributedLocker(NewMemoryDriver())
	ctx := context.Background()

	lock1, err := locker.Acquire(ctx, "exclusive_key", WithDistLockTTL(1*time.Second), WithDistLockOwner("worker_A"))
	if err != nil {
		t.Fatalf("worker A failed to acquire lock: %v", err)
	}
	defer lock1.Release()

	// Worker B attempts to acquire same key -> must fail immediately
	_, err2 := locker.Acquire(ctx, "exclusive_key", WithDistLockTTL(1*time.Second), WithDistLockOwner("worker_B"))
	if !errors.Is(err2, ErrLockNotAcquired) {
		t.Fatalf("expected ErrLockNotAcquired for worker B, got %v", err2)
	}
}

func TestDistributedLocker_MonotonicFencingTokens(t *testing.T) {
	locker := NewDistributedLocker(NewMemoryDriver())
	ctx := context.Background()

	lock1, err := locker.Acquire(ctx, "shared_job", WithDistLockTTL(1*time.Second))
	if err != nil {
		t.Fatalf("lock 1 failed: %v", err)
	}
	token1 := lock1.FencingToken()
	_ = lock1.Release()

	lock2, err := locker.Acquire(ctx, "shared_job", WithDistLockTTL(1*time.Second))
	if err != nil {
		t.Fatalf("lock 2 failed: %v", err)
	}
	token2 := lock2.FencingToken()
	_ = lock2.Release()

	if token2 <= token1 {
		t.Fatalf("fencing tokens must be strictly monotonic: token1=%d token2=%d", token1, token2)
	}
}

func TestDistributedLocker_AutoHeartbeatRenewal(t *testing.T) {
	locker := NewDistributedLocker(NewMemoryDriver())
	ctx := context.Background()

	// TTL is 30ms, but AutoRenew refreshes it every 10ms
	lock, err := locker.Acquire(ctx, "long_task", WithDistLockTTL(30*time.Millisecond), WithDistLockAutoRenew(true))
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	defer lock.Release()

	// Sleep for 70ms (well past the 30ms initial TTL)
	time.Sleep(70 * time.Millisecond)

	// Another worker should STILL not be able to acquire because heartbeat extended it!
	_, err2 := locker.Acquire(ctx, "long_task", WithDistLockOwner("worker_other"))
	if !errors.Is(err2, ErrLockNotAcquired) {
		t.Fatalf("expected lock to remain held due to heartbeat renewal, got: %v", err2)
	}
}

func TestDistributedLocker_RetryAcquire(t *testing.T) {
	locker := NewDistributedLocker(NewMemoryDriver())
	ctx := context.Background()

	lock1, _ := locker.Acquire(ctx, "retry_key", WithDistLockTTL(1*time.Second), WithDistLockOwner("first"))

	var wg sync.WaitGroup
	wg.Add(1)

	var lock2 *DistLock
	var err2 error

	go func() {
		defer wg.Done()
		// Will wait up to 500ms with 10ms polls
		lock2, err2 = locker.Acquire(ctx, "retry_key",
			WithDistLockOwner("second"),
			WithDistLockRetry(500*time.Millisecond, 10*time.Millisecond),
		)
	}()

	// Release lock1 after 50ms
	time.Sleep(50 * time.Millisecond)
	_ = lock1.Release()

	wg.Wait()

	if err2 != nil {
		t.Fatalf("expected second worker to acquire after release, got %v", err2)
	}
	if lock2 == nil || lock2.Owner() != "second" {
		t.Fatalf("unexpected lock2 state: %+v", lock2)
	}
	_ = lock2.Release()
}
