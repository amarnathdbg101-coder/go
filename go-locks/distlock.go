package golocks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
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

// DistLock represents an actively held distributed lease with fencing guarantees.
type DistLock struct {
	key          string
	owner        string
	fencingToken uint64
	ttl          time.Duration
	driver       Driver

	mu        sync.Mutex
	renewStop chan struct{}
	released  bool
}

// FencingToken returns the strictly monotonic sequence token issued for this lock acquisition.
// Use this token in storage updates (e.g. UPDATE ... WHERE fencing_token < :token) to prevent split-brain writes.
func (l *DistLock) FencingToken() uint64 {
	return l.fencingToken
}

// Key returns the locked resource key.
func (l *DistLock) Key() string {
	return l.key
}

// Owner returns the unique worker ID holding this lock.
func (l *DistLock) Owner() string {
	return l.owner
}

// Release releases the lock and terminates the background auto-renewal heartbeat.
func (l *DistLock) Release() error {
	l.mu.Lock()
	if l.released {
		l.mu.Unlock()
		return ErrLockReleased
	}
	l.released = true
	if l.renewStop != nil {
		close(l.renewStop)
	}
	driver := l.driver
	key := l.key
	owner := l.owner
	l.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return driver.Unlock(ctx, key, owner)
}

// DistLockConfig specifies the acquisition and lease renewal parameters for a distributed lock.
type DistLockConfig struct {
	TTL           time.Duration
	AutoRenew     bool
	RenewInterval time.Duration
	RetryWait     time.Duration
	RetryTimeout  time.Duration
	Owner         string
}

// DefaultDistLockConfig returns production defaults for distributed locks.
func DefaultDistLockConfig() DistLockConfig {
	return DistLockConfig{
		TTL:          5 * time.Second,
		AutoRenew:    true,
		RetryWait:    50 * time.Millisecond,
		RetryTimeout: 0,
	}
}

// DistLockOption is a functional option for configuring distributed locks.
type DistLockOption func(*DistLockConfig)

// WithDistLockTTL sets the lock expiration lease time.
func WithDistLockTTL(d time.Duration) DistLockOption {
	return func(c *DistLockConfig) {
		if d > 0 {
			c.TTL = d
		}
	}
}

// WithDistLockAutoRenew toggles automatic heartbeat lease renewal.
func WithDistLockAutoRenew(enable bool) DistLockOption {
	return func(c *DistLockConfig) {
		c.AutoRenew = enable
	}
}

// WithDistLockRetry sets polling retry parameters for waiting on a locked resource.
func WithDistLockRetry(timeout, wait time.Duration) DistLockOption {
	return func(c *DistLockConfig) {
		if timeout > 0 {
			c.RetryTimeout = timeout
		}
		if wait > 0 {
			c.RetryWait = wait
		}
	}
}

// WithDistLockOwner overrides the worker owner identifier.
func WithDistLockOwner(owner string) DistLockOption {
	return func(c *DistLockConfig) {
		c.Owner = owner
	}
}

// DistributedLocker coordinates distributed lock acquisition and heartbeats.
type DistributedLocker struct {
	driver Driver
}

// NewDistributedLocker initializes a distributed lock manager backed by the specified driver.
func NewDistributedLocker(driver Driver) *DistributedLocker {
	if driver == nil {
		driver = NewMemoryDriver()
	}
	return &DistributedLocker{driver: driver}
}

// Acquire attempts to acquire the named lock adhering to the specified options.
func (lk *DistributedLocker) Acquire(ctx context.Context, key string, opts ...DistLockOption) (*DistLock, error) {
	cfg := DefaultDistLockConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Owner == "" {
		cfg.Owner = generateDistLockOwnerID()
	}
	if cfg.RenewInterval <= 0 {
		cfg.RenewInterval = cfg.TTL / 3
	}

	deadline := time.Now().Add(cfg.RetryTimeout)
	for {
		token, err := lk.driver.TryLock(ctx, key, cfg.Owner, cfg.TTL)
		if err == nil {
			lock := &DistLock{
				key:          key,
				owner:        cfg.Owner,
				fencingToken: token,
				ttl:          cfg.TTL,
				driver:       lk.driver,
			}

			// Launch auto-renewal heartbeat if requested
			if cfg.AutoRenew {
				lock.renewStop = make(chan struct{})
				go lock.startHeartbeat(cfg.RenewInterval)
			}

			return lock, nil
		}

		if cfg.RetryTimeout <= 0 || time.Now().After(deadline) {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(cfg.RetryWait):
		}
	}
}

func (l *DistLock) startHeartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-l.renewStop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), interval)
			err := l.driver.Extend(ctx, l.key, l.owner, l.ttl)
			cancel()
			if err != nil {
				// Lease lost or expired
				return
			}
		}
	}
}

func generateDistLockOwnerID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("worker_%s", hex.EncodeToString(b[:]))
}
