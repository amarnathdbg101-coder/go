package distlock

import "time"

// Config specifies the acquisition and lease renewal parameters for a distributed lock.
type Config struct {
	// TTL is the lease time before the lock automatically expires.
	// Default is 5 seconds.
	TTL time.Duration

	// AutoRenew enables a background heartbeat goroutine to extend the lease until explicitly released.
	// Default is true.
	AutoRenew bool

	// RenewInterval specifies how frequently to refresh the lease.
	// Default is TTL / 3.
	RenewInterval time.Duration

	// RetryWait is the pause between acquisition attempts if RetryTimeout > 0.
	// Default is 50 milliseconds.
	RetryWait time.Duration

	// RetryTimeout is the maximum duration to wait/poll for an unavailable lock.
	// Default is 0 (fail-fast immediately if lock is unavailable).
	RetryTimeout time.Duration

	// Owner is the unique worker identifier holding the lock.
	Owner string
}

// Option is a functional option for configuring distributed locks.
type Option func(*Config)

// DefaultConfig returns production defaults for distributed locks.
func DefaultConfig() Config {
	return Config{
		TTL:          5 * time.Second,
		AutoRenew:    true,
		RetryWait:    50 * time.Millisecond,
		RetryTimeout: 0,
	}
}

// WithTTL sets the lock expiration lease time.
func WithTTL(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.TTL = d
		}
	}
}

// WithAutoRenew toggles automatic heartbeat lease renewal.
func WithAutoRenew(enable bool) Option {
	return func(c *Config) {
		c.AutoRenew = enable
	}
}

// WithRetry sets polling retry parameters for waiting on a locked resource.
func WithRetry(timeout, wait time.Duration) Option {
	return func(c *Config) {
		if timeout > 0 {
			c.RetryTimeout = timeout
		}
		if wait > 0 {
			c.RetryWait = wait
		}
	}
}

// WithOwner overrides the worker owner identifier.
func WithOwner(owner string) Option {
	return func(c *Config) {
		c.Owner = owner
	}
}
