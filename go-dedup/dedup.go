package dedup

import (
	"sync"
	"time"
)

// Deduplicator manages sliding-window deduplication for arbitrary comparable keys.
type Deduplicator[K comparable] struct {
	mu            sync.RWMutex
	items         map[K]int64 // Key -> UnixNano expiration timestamp
	defaultWindow time.Duration
	stopChan      chan struct{}
	closed        bool
}

// New creates a new Deduplicator instance with a default retention window and background cleanup interval.
// If cleanupInterval is 0, automatic background purging is disabled.
func New[K comparable](defaultWindow time.Duration, cleanupInterval time.Duration) *Deduplicator[K] {
	d := &Deduplicator[K]{
		items:         make(map[K]int64),
		defaultWindow: defaultWindow,
		stopChan:      make(chan struct{}),
	}

	if cleanupInterval > 0 {
		go d.startJanitor(cleanupInterval)
	}

	return d
}

// CheckAndRecord atomically checks if a key is a duplicate within the default retention window.
// Returns true if the key was already present and not expired (duplicate detected).
// Returns false if the key was new or expired (successfully recorded).
func (d *Deduplicator[K]) CheckAndRecord(key K) bool {
	return d.CheckAndRecordWithTTL(key, d.defaultWindow)
}

// CheckAndRecordWithTTL atomically checks if a key is a duplicate with a custom TTL.
// Returns true if duplicate, false if newly recorded.
func (d *Deduplicator[K]) CheckAndRecordWithTTL(key K, ttl time.Duration) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return false
	}

	now := time.Now().UnixNano()
	if exp, exists := d.items[key]; exists {
		if exp > now {
			return true // Is duplicate
		}
	}

	// Record or refresh
	d.items[key] = now + ttl.Nanoseconds()
	return false
}

// IsDuplicate returns true if the key is currently active and not expired without recording or extending TTL.
func (d *Deduplicator[K]) IsDuplicate(key K) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return false
	}

	exp, exists := d.items[key]
	if !exists {
		return false
	}

	return exp > time.Now().UnixNano()
}

// Record forces recording or extending a key with the default retention window.
func (d *Deduplicator[K]) Record(key K) {
	d.RecordWithTTL(key, d.defaultWindow)
}

// RecordWithTTL forces recording or extending a key with a specific TTL.
func (d *Deduplicator[K]) RecordWithTTL(key K, ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	d.items[key] = time.Now().UnixNano() + ttl.Nanoseconds()
}

// Remove explicitly invalidates and deletes a key from deduplication storage.
func (d *Deduplicator[K]) Remove(key K) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.items, key)
}

// Count returns the number of currently active, unexpired keys.
func (d *Deduplicator[K]) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	now := time.Now().UnixNano()
	count := 0
	for _, exp := range d.items {
		if exp > now {
			count++
		}
	}
	return count
}

// Clear removes all keys from the deduplicator.
func (d *Deduplicator[K]) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.items = make(map[K]int64)
}

// PurgeExpired synchronously deletes all expired keys from memory.
func (d *Deduplicator[K]) PurgeExpired() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UnixNano()
	evicted := 0
	for k, exp := range d.items {
		if exp <= now {
			delete(d.items, k)
			evicted++
		}
	}
	return evicted
}

// Close stops the background janitor goroutine and clears the deduplicator.
func (d *Deduplicator[K]) Close() {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return
	}
	d.closed = true
	close(d.stopChan)
	d.items = nil
	d.mu.Unlock()
}

func (d *Deduplicator[K]) startJanitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.PurgeExpired()
		case <-d.stopChan:
			return
		}
	}
}
