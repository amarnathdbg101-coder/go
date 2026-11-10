package distlock

import (
	"context"
	"sync"
	"time"
)

type memoryLockEntry struct {
	owner        string
	expiresAt    time.Time
	fencingToken uint64
}

// MemoryDriver provides an in-memory driver with monotonic fencing tokens and TTL expiration.
type MemoryDriver struct {
	mu             sync.Mutex
	fencingCounter uint64
	locks          map[string]*memoryLockEntry
}

// NewMemoryDriver creates a thread-safe in-memory distributed lock driver.
func NewMemoryDriver() *MemoryDriver {
	return &MemoryDriver{
		locks: make(map[string]*memoryLockEntry),
	}
}

// TryLock attempts to acquire the lock.
func (d *MemoryDriver) TryLock(ctx context.Context, key, owner string, ttl time.Duration) (uint64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	entry, exists := d.locks[key]

	if exists && now.Before(entry.expiresAt) {
		if entry.owner != owner {
			return 0, ErrLockNotAcquired
		}
		// Same owner re-acquiring: update lease
		entry.expiresAt = now.Add(ttl)
		return entry.fencingToken, nil
	}

	// Lock is free or expired: grant lock with next monotonic fencing token
	d.fencingCounter++
	token := d.fencingCounter

	d.locks[key] = &memoryLockEntry{
		owner:        owner,
		expiresAt:    now.Add(ttl),
		fencingToken: token,
	}

	return token, nil
}

// Unlock releases the lock if owner matches.
func (d *MemoryDriver) Unlock(ctx context.Context, key, owner string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry, exists := d.locks[key]
	if !exists || time.Now().After(entry.expiresAt) {
		delete(d.locks, key)
		return nil
	}

	if entry.owner != owner {
		return ErrOwnerMismatch
	}

	delete(d.locks, key)
	return nil
}

// Extend refreshes the lock lease time.
func (d *MemoryDriver) Extend(ctx context.Context, key, owner string, ttl time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry, exists := d.locks[key]
	if !exists || time.Now().After(entry.expiresAt) {
		return ErrLockExpired
	}

	if entry.owner != owner {
		return ErrOwnerMismatch
	}

	entry.expiresAt = time.Now().Add(ttl)
	return nil
}

// Close releases the driver map.
func (d *MemoryDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.locks = nil
	return nil
}
