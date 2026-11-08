package taskqueue

import (
	"runtime"
	"time"
)

// Config holds runtime options for the TaskQueue worker pool.
type Config struct {
	// Workers is the fixed number of concurrent worker goroutines processing tasks.
	// Default: runtime.NumCPU() * 2 (min 4).
	Workers int

	// BufferSize defines the capacity of the in-memory buffered channel before blocking or returning ErrQueueFull.
	// Default: 1000.
	BufferSize int

	// MaxRetries is the maximum number of times a failing task will be retried with exponential backoff.
	// Default: 3.
	MaxRetries int

	// InitialBackoff is the starting delay before the first retry attempt.
	// Default: 500ms.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum delay cap for exponential backoff retries.
	// Default: 30s.
	MaxBackoff time.Duration

	// TaskTimeout is the maximum time a single task is allowed to execute before being cancelled.
	// Default: 30s.
	TaskTimeout time.Duration
}

// Option customizes Config settings.
type Option func(*Config)

// DefaultConfig provides production-ready worker pool defaults.
func DefaultConfig() Config {
	workers := runtime.NumCPU() * 2
	if workers < 4 {
		workers = 4
	}

	return Config{
		Workers:        workers,
		BufferSize:     1000,
		MaxRetries:     3,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     30 * time.Second,
		TaskTimeout:    30 * time.Second,
	}
}

// WithWorkers sets worker pool concurrency.
func WithWorkers(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.Workers = n
		}
	}
}

// WithBufferSize sets channel queue buffer capacity.
func WithBufferSize(size int) Option {
	return func(c *Config) {
		if size > 0 {
			c.BufferSize = size
		}
	}
}

// WithMaxRetries sets retry limit for failing tasks.
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		if retries >= 0 {
			c.MaxRetries = retries
		}
	}
}

// WithTaskTimeout sets individual task execution deadline.
func WithTaskTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if timeout > 0 {
			c.TaskTimeout = timeout
		}
	}
}

// WithBackoff configures initial and maximum exponential retry backoff intervals.
func WithBackoff(initial, max time.Duration) Option {
	return func(c *Config) {
		if initial > 0 {
			c.InitialBackoff = initial
		}
		if max > 0 {
			c.MaxBackoff = max
		}
	}
}
