package pgexplain

import "time"

// Config defines the thresholds and alert callbacks for PostgreSQL query inspection.
type Config struct {
	// SlowThreshold is the query execution duration threshold that triggers automatic EXPLAIN.
	// Default is 100 milliseconds.
	SlowThreshold time.Duration

	// OnSlowQuery is the diagnostic callback executed when an execution plan is generated.
	OnSlowQuery func(report *Report)

	// AsyncExplain specifies whether to run the EXPLAIN query asynchronously to avoid blocking callers.
	// Default is true.
	AsyncExplain bool
}

// Option is a functional option for configuring the pgexplain auditor.
type Option func(*Config)

// DefaultConfig returns production defaults for PostgreSQL query inspection.
func DefaultConfig() Config {
	return Config{
		SlowThreshold: 100 * time.Millisecond,
		AsyncExplain:  true,
		OnSlowQuery: func(r *Report) {
			// Fallback no-op
		},
	}
}

// WithSlowThreshold sets the threshold for triggering EXPLAIN analysis.
func WithSlowThreshold(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.SlowThreshold = d
		}
	}
}

// WithOnSlowQuery registers a diagnostic report callback.
func WithOnSlowQuery(fn func(report *Report)) Option {
	return func(c *Config) {
		c.OnSlowQuery = fn
	}
}

// WithAsyncExplain configures whether EXPLAIN runs in a non-blocking background goroutine.
func WithAsyncExplain(async bool) Option {
	return func(c *Config) {
		c.AsyncExplain = async
	}
}
