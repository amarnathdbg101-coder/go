package distlock

import (
	"context"
	"time"
)

// Driver abstracts the distributed storage backend (Redis, Postgres, or In-Memory).
type Driver interface {
	// TryLock attempts to acquire the lock. If successful, returns the monotonically increasing fencing token.
	// If already held by another owner, returns ErrLockNotAcquired.
	TryLock(ctx context.Context, key, owner string, ttl time.Duration) (uint64, error)

	// Unlock releases the lock if and only if it is owned by the specified owner.
	Unlock(ctx context.Context, key, owner string) error

	// Extend renews the TTL of the lock if owned by the specified owner.
	Extend(ctx context.Context, key, owner string, ttl time.Duration) error

	// Close releases any driver resources.
	Close() error
}
