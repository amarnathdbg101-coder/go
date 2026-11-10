package trafficshaper

import "time"

// Algorithm specifies the traffic shaping model.
type Algorithm int

const (
	// AlgorithmLeakyBucket strictly paces emissions at a constant rate with zero burst tolerance.
	AlgorithmLeakyBucket Algorithm = iota
	// AlgorithmTokenBucket permits temporary bursts up to bucket capacity while sustaining average rate.
	AlgorithmTokenBucket
)

// DropPolicy defines behavior when queue capacity is exceeded.
type DropPolicy int

const (
	// DropPolicyReject immediately returns ErrQueueFull to caller.
	DropPolicyReject DropPolicy = iota
	// DropPolicyDropTail drops the incoming element silently or reports dropped metric.
	DropPolicyDropTail
)

// Config configures the traffic shaper engine.
type Config struct {
	Algorithm     Algorithm
	Rate          float64       // Target requests per second
	Burst         int           // Maximum burst size (for token bucket)
	QueueCapacity int           // Maximum items waiting in shaper queue
	MaxWaitTime   time.Duration // Maximum wait time before timing out
	DropPolicy    DropPolicy
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig returns recommended production defaults.
func DefaultConfig() Config {
	return Config{
		Algorithm:     AlgorithmTokenBucket,
		Rate:          100.0,
		Burst:         20,
		QueueCapacity: 500,
		MaxWaitTime:   5 * time.Second,
		DropPolicy:    DropPolicyReject,
	}
}

// WithAlgorithm sets the traffic shaping algorithm.
func WithAlgorithm(algo Algorithm) Option {
	return func(c *Config) {
		c.Algorithm = algo
	}
}

// WithRate sets the target continuous throughput in operations per second.
func WithRate(rate float64) Option {
	return func(c *Config) {
		c.Rate = rate
	}
}

// WithBurst sets the burst capacity.
func WithBurst(burst int) Option {
	return func(c *Config) {
		c.Burst = burst
	}
}

// WithQueueCapacity sets the maximum waiting queue capacity.
func WithQueueCapacity(cap int) Option {
	return func(c *Config) {
		c.QueueCapacity = cap
	}
}

// WithMaxWaitTime sets the queue deadline before returning timeout.
func WithMaxWaitTime(d time.Duration) Option {
	return func(c *Config) {
		c.MaxWaitTime = d
	}
}

// WithDropPolicy sets the policy when buffer capacity overflows.
func WithDropPolicy(p DropPolicy) Option {
	return func(c *Config) {
		c.DropPolicy = p
	}
}
