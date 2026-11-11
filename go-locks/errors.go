package golocks

import "errors"

var (
	// ErrLockNotAcquired indicates that TryLock failed to acquire the lock.
	ErrLockNotAcquired = errors.New("locks: failed to acquire lock (already held)")

	// ErrLockTimeout indicates the acquisition context timed out.
	ErrLockTimeout = errors.New("locks: lock acquisition timed out")

	// ErrLockNotHeld indicates attempting to unlock a lock not owned or already released.
	ErrLockNotHeld = errors.New("locks: lock is not held")

	// ErrLockExpired indicates the operation failed because the lease TTL expired.
	ErrLockExpired = errors.New("locks: lock lease has expired")

	// ErrOwnerMismatch indicates an attempt to unlock or extend a lock held by a different owner.
	ErrOwnerMismatch = errors.New("locks: cannot release or extend lock owned by different worker")

	// ErrLockReleased indicates an operation was attempted on an already released lock.
	ErrLockReleased = errors.New("locks: lock is already released")
)
