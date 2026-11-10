package distlock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Lock represents an actively held distributed lease with fencing guarantees.
type Lock struct {
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
func (l *Lock) FencingToken() uint64 {
	return l.fencingToken
}

// Key returns the locked resource key.
func (l *Lock) Key() string {
	return l.key
}

// Owner returns the unique worker ID holding this lock.
func (l *Lock) Owner() string {
	return l.owner
}

// Release releases the lock and terminates the background auto-renewal heartbeat.
func (l *Lock) Release() error {
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

// Locker coordinates distributed lock acquisition and heartbeats.
type Locker struct {
	driver Driver
}

var defaultLocker = NewLocker(NewMemoryDriver())

// NewLocker initializes a distributed lock manager backed by the specified driver.
func NewLocker(driver Driver) *Locker {
	if driver == nil {
		driver = NewMemoryDriver()
	}
	return &Locker{driver: driver}
}

// Acquire acquires a distributed lock using the default locker.
func Acquire(ctx context.Context, key string, opts ...Option) (*Lock, error) {
	return defaultLocker.Acquire(ctx, key, opts...)
}

// Acquire attempts to acquire the named lock adhering to the specified options.
func (lk *Locker) Acquire(ctx context.Context, key string, opts ...Option) (*Lock, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Owner == "" {
		cfg.Owner = generateOwnerID()
	}
	if cfg.RenewInterval <= 0 {
		cfg.RenewInterval = cfg.TTL / 3
	}

	deadline := time.Now().Add(cfg.RetryTimeout)
	for {
		token, err := lk.driver.TryLock(ctx, key, cfg.Owner, cfg.TTL)
		if err == nil {
			lock := &Lock{
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

func (l *Lock) startHeartbeat(interval time.Duration) {
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

func generateOwnerID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("worker_%s", hex.EncodeToString(b[:]))
}
