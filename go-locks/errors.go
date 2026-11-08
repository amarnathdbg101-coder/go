package golocks

import "errors"

var (
	// ErrLockNotAcquired indicates that TryLock failed to acquire the lock.
	ErrLockNotAcquired = errors.New("locks: failed to acquire lock (already held)")

	// ErrLockTimeout indicates the acquisition context timed out.
	ErrLockTimeout = errors.New("locks: lock acquisition timed out")

	// ErrLockNotHeld indicates attempting to unlock a lock not owned or already released.
	ErrLockNotHeld = errors.New("locks: lock is not held")
)
