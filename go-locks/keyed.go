package golocks

import (
	"sync"
)

type refMutex struct {
	mu       sync.Mutex
	refCount int
}

// KeyedMutex provides granular per-key locking so operations on different keys proceed concurrently.
// Uses reference counting to clean up unused keys automatically, preventing memory leaks.
type KeyedMutex struct {
	mu    sync.Mutex
	locks map[string]*refMutex
}

// NewKeyedMutex constructs an empty KeyedMutex.
func NewKeyedMutex() *KeyedMutex {
	return &KeyedMutex{
		locks: make(map[string]*refMutex),
	}
}

// Lock acquires a lock for key and returns an unlock function.
func (km *KeyedMutex) Lock(key string) (unlock func()) {
	km.mu.Lock()
	entry, ok := km.locks[key]
	if !ok {
		entry = &refMutex{}
		km.locks[key] = entry
	}
	entry.refCount++
	km.mu.Unlock()

	// Acquire actual key mutex
	entry.mu.Lock()

	return func() {
		entry.mu.Unlock()

		km.mu.Lock()
		defer km.mu.Unlock()

		entry.refCount--
		if entry.refCount <= 0 {
			delete(km.locks, key)
		}
	}
}

// TryLock attempts to acquire the lock for key without blocking.
func (km *KeyedMutex) TryLock(key string) (unlock func(), acquired bool) {
	km.mu.Lock()
	entry, ok := km.locks[key]
	if !ok {
		entry = &refMutex{}
		km.locks[key] = entry
	}

	// Try acquiring
	if !entry.mu.TryLock() {
		// If entry was just created and we failed to lock it (impossible unless raced), don't leak
		if entry.refCount == 0 {
			delete(km.locks, key)
		}
		km.mu.Unlock()
		return nil, false
	}

	entry.refCount++
	km.mu.Unlock()

	return func() {
		entry.mu.Unlock()

		km.mu.Lock()
		defer km.mu.Unlock()

		entry.refCount--
		if entry.refCount <= 0 {
			delete(km.locks, key)
		}
	}, true
}

// WithLock executes fn while holding the lock for key.
func (km *KeyedMutex) WithLock(key string, fn func() error) error {
	unlock := km.Lock(key)
	defer unlock()
	return fn()
}

// ActiveKeys returns count of currently tracked keys (for metrics/debugging).
func (km *KeyedMutex) ActiveKeys() int {
	km.mu.Lock()
	defer km.mu.Unlock()
	return len(km.locks)
}
