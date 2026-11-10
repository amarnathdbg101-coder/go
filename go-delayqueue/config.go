package delayqueue

import "time"

// JobStatus represents the lifecycle state of a scheduled job.
type JobStatus string

const (
	StatusPending    JobStatus = "PENDING"
	StatusProcessing JobStatus = "PROCESSING"
	StatusCompleted  JobStatus = "COMPLETED"
	StatusFailed     JobStatus = "FAILED"
	StatusDeadLetter JobStatus = "DEAD_LETTER"
)

// Job represents a delayed unit of work.
type Job struct {
	ID             string    `json:"id"`
	Topic          string    `json:"topic"`
	Payload        []byte    `json:"payload"`
	ExecuteAt      time.Time `json:"execute_at"`
	MaxRetries     int       `json:"max_retries"`
	CurrentRetries int       `json:"current_retries"`
	Status         JobStatus `json:"status"`
	LastError      string    `json:"last_error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Config specifies the runtime settings for the delay queue engine.
type Config struct {
	// PollInterval is the polling interval to check for due jobs.
	// Default is 50 milliseconds.
	PollInterval time.Duration

	// BatchSize is the max jobs to pull in a single dispatch cycle.
	// Default is 50.
	BatchSize int

	// WorkerCount is the number of concurrent worker goroutines.
	// Default is 4.
	WorkerCount int

	// DefaultMaxRetries is the fallback retry count if not specified per job.
	// Default is 3.
	DefaultMaxRetries int

	// RetryBackoff is the base backoff interval for retries.
	// Default is 1 second.
	RetryBackoff time.Duration
}

// Option is a functional option for configuring the delay queue.
type Option func(*Config)

// DefaultConfig returns production defaults for the delay queue.
func DefaultConfig() Config {
	return Config{
		PollInterval:      50 * time.Millisecond,
		BatchSize:         50,
		WorkerCount:       4,
		DefaultMaxRetries: 3,
		RetryBackoff:      1 * time.Second,
	}
}

// WithPollInterval sets the background polling interval.
func WithPollInterval(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.PollInterval = d
		}
	}
}

// WithBatchSize sets the max batch size of due jobs.
func WithBatchSize(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.BatchSize = n
		}
	}
}

// WithWorkerCount sets the number of concurrent worker goroutines.
func WithWorkerCount(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.WorkerCount = n
		}
	}
}

// WithDefaultMaxRetries sets default retries per job.
func WithDefaultMaxRetries(n int) Option {
	return func(c *Config) {
		if n >= 0 {
			c.DefaultMaxRetries = n
		}
	}
}

// WithRetryBackoff sets the base retry backoff interval.
func WithRetryBackoff(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.RetryBackoff = d
		}
	}
}
