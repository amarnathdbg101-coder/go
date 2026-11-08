package golocks

import (
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
