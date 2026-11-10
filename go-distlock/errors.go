package distlock

import "errors"

var (
	// ErrLockNotAcquired is returned when the requested lock is already held by another owner.
	ErrLockNotAcquired = errors.New("distlock: lock is currently held by another worker")

	// ErrLockExpired is returned when a lock has expired or lease extension failed.
	ErrLockExpired = errors.New("distlock: lock has expired or could not be renewed")

	// ErrLockReleased is returned when attempting an operation on an already released lock.
	ErrLockReleased = errors.New("distlock: lock has already been released")

	// ErrOwnerMismatch is returned when unlocking a lock owned by a different client.
	ErrOwnerMismatch = errors.New("distlock: owner mismatch, cannot release or extend foreign lock")
)
