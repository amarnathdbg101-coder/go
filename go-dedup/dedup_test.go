package dedup

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDeduplicator_CheckAndRecord(t *testing.T) {
	d := New[string](100*time.Millisecond, 0)
	defer d.Close()

	key := "event-12345"

	// First occurrence: should NOT be duplicate
	if d.CheckAndRecord(key) {
		t.Errorf("expected first occurrence to be non-duplicate (false)")
	}

	// Immediate repeat: SHOULD be duplicate
	if !d.CheckAndRecord(key) {
		t.Errorf("expected second occurrence to be duplicate (true)")
	}

	// Non-mutating check
	if !d.IsDuplicate(key) {
		t.Errorf("expected IsDuplicate to be true")
	}

	// Wait for TTL expiration
	time.Sleep(120 * time.Millisecond)

	// After TTL: should NOT be duplicate again
	if d.CheckAndRecord(key) {
		t.Errorf("expected occurrence after TTL to be non-duplicate (false)")
	}
}

func TestDeduplicator_CustomTTL(t *testing.T) {
	d := New[string](10*time.Second, 0)
	defer d.Close()

	key := "quick-event"
	// 50ms TTL instead of 10s
	if d.CheckAndRecordWithTTL(key, 50*time.Millisecond) {
		t.Errorf("expected initial check to be false")
	}

	if !d.IsDuplicate(key) {
		t.Errorf("expected key to be duplicate")
	}

	time.Sleep(70 * time.Millisecond)

	if d.IsDuplicate(key) {
		t.Errorf("expected key to have expired")
	}
}

func TestDeduplicator_RemoveAndClear(t *testing.T) {
	d := New[int](1*time.Minute, 0)
	defer d.Close()

	d.Record(1)
	d.Record(2)
	d.Record(3)

	if d.Count() != 3 {
		t.Errorf("expected count 3, got %d", d.Count())
	}

	d.Remove(2)
	if d.IsDuplicate(2) {
		t.Errorf("expected key 2 to be removed")
	}
	if d.Count() != 2 {
		t.Errorf("expected count 2, got %d", d.Count())
	}

	d.Clear()
	if d.Count() != 0 {
		t.Errorf("expected count 0 after clear, got %d", d.Count())
	}
}

func TestDeduplicator_PurgeExpired(t *testing.T) {
	// Disable background janitor to test synchronous PurgeExpired()
	d := New[string](30*time.Millisecond, 0)
	defer d.Close()

	d.Record("e1")
	d.Record("e2")
	d.RecordWithTTL("e3", 500*time.Millisecond)

	// Sleep 80ms to provide ample margin over 30ms TTL under heavy CPU load
	time.Sleep(80 * time.Millisecond)

	// e1 and e2 should have expired, e3 still valid
	evicted := d.PurgeExpired()
	if evicted != 2 {
		t.Errorf("expected 2 evicted keys, got %d", evicted)
	}

	if d.Count() != 1 {
		t.Errorf("expected 1 remaining key (e3), got %d", d.Count())
	}
}

func TestDeduplicator_BackgroundJanitor(t *testing.T) {
	// Enable background janitor every 25ms
	d := New[string](40*time.Millisecond, 25*time.Millisecond)
	defer d.Close()

	d.Record("auto1")
	d.Record("auto2")

	// Wait for TTL + janitor run
	time.Sleep(100 * time.Millisecond)

	d.mu.RLock()
	rawLen := len(d.items)
	d.mu.RUnlock()

	if rawLen != 0 {
		t.Errorf("expected janitor to purge all items, map still has %d items", rawLen)
	}
}

func TestDeduplicator_ConcurrentAccess(t *testing.T) {
	d := New[string](1*time.Minute, 0)
	defer d.Close()

	const workers = 50
	const iterations = 50
	const distinctKeys = 10

	var recordedCount int64
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				key := fmt.Sprintf("key-%d", i%distinctKeys)
				isDup := d.CheckAndRecord(key)
				if !isDup {
					atomic.AddInt64(&recordedCount, 1)
				}
			}
		}(w)
	}

	wg.Wait()

	// Exactly distinctKeys should have been successfully recorded as non-duplicates
	if recordedCount != int64(distinctKeys) {
		t.Errorf("expected exactly %d unique records, got %d", distinctKeys, recordedCount)
	}
}
