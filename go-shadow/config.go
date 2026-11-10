package shadow

import "time"

// DiffReport details discrepancies between primary and candidate responses.
type DiffReport struct {
	Method          string        `json:"method"`
	Path            string        `json:"path"`
	PrimaryStatus   int           `json:"primary_status"`
	CandidateStatus int           `json:"candidate_status"`
	StatusMismatch  bool          `json:"status_mismatch"`
	BodyMismatch    bool          `json:"body_mismatch"`
	PrimaryLatency  time.Duration `json:"primary_latency"`
	CandidateLatency time.Duration `json:"candidate_latency"`
	PrimaryBody     string        `json:"primary_body,omitempty"`
	CandidateBody   string        `json:"candidate_body,omitempty"`
	Timestamp       time.Time     `json:"timestamp"`
}

// Config controls the behavior of the traffic shadowing engine.
type Config struct {
	// TargetURL is the candidate service base URL (e.g. "http://candidate-v2:8081").
	TargetURL string

	// SampleRate is the fraction of traffic to mirror (0.0 to 1.0). Default is 1.0 (100%).
	SampleRate float64

	// Timeout is the maximum duration to wait for candidate responses. Default is 3 seconds.
	Timeout time.Duration

	// CompareBody toggles response body comparison. Default is true.
	CompareBody bool

	// OnDiff is triggered whenever a difference in status code or body is detected.
	OnDiff func(diff *DiffReport)

	// MaxWorkerQueue caps the background shadow dispatch queue to prevent memory growth.
	MaxWorkerQueue int
}

// Option is a functional option for configuring the shadow engine.
type Option func(*Config)

// DefaultConfig returns production defaults for traffic shadowing.
func DefaultConfig() Config {
	return Config{
		SampleRate:     1.0,
		Timeout:        3 * time.Second,
		CompareBody:    true,
		MaxWorkerQueue: 500,
	}
}

// WithTargetURL specifies the destination service for shadowed traffic.
func WithTargetURL(target string) Option {
	return func(c *Config) {
		c.TargetURL = target
	}
}

// WithSampleRate sets the mirroring percentage (e.g. 0.20 = 20%).
func WithSampleRate(rate float64) Option {
	return func(c *Config) {
		if rate >= 0 && rate <= 1.0 {
			c.SampleRate = rate
		}
	}
}

// WithTimeout sets the candidate request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.Timeout = d
		}
	}
}

// WithOnDiff registers a callback for detected discrepancies.
func WithOnDiff(fn func(diff *DiffReport)) Option {
	return func(c *Config) {
		c.OnDiff = fn
	}
}
