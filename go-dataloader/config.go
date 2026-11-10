package dataloader

import "time"

// Result encapsulates a loaded value or an error for an individual key in a batch.
type Result[V any] struct {
	Data V
	Err  error
}

// Config holds tuneable settings for a DataLoader instance.
type Config struct {
	// Wait is the debounce window during which keys are collected into a batch before triggering.
	// Default is 2 milliseconds.
	Wait time.Duration

	// MaxBatch is the maximum number of keys to include in a single batch call.
	// Default is 100.
	MaxBatch int

	// Cache enables or disables per-loader memoization cache.
	// Default is true.
	Cache bool
}

// Option is a functional option for configuring DataLoader.
type Option func(*Config)

// DefaultConfig returns optimal production defaults for DataLoader.
func DefaultConfig() Config {
	return Config{
		Wait:     2 * time.Millisecond,
		MaxBatch: 100,
		Cache:    true,
	}
}

// WithWait sets the batching collection window.
func WithWait(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.Wait = d
		}
	}
}

// WithMaxBatch sets the upper bound on keys per single batch dispatch.
func WithMaxBatch(max int) Option {
	return func(c *Config) {
		if max > 0 {
			c.MaxBatch = max
		}
	}
}

// WithCache toggles internal memoization caching.
func WithCache(enable bool) Option {
	return func(c *Config) {
		c.Cache = enable
	}
}
