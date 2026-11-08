package keyrotator

import "time"

// Option configures KeyRotator parameters.
type Option func(*Config)

// Hooks defines zero-dependency observability callbacks for metrics, tracing, and logging.
type Hooks struct {
	// OnKeySelected is invoked whenever a healthy key is retrieved for execution.
	OnKeySelected func(key string)

	// OnRateLimited is invoked when a key enters rate-limit quarantine (e.g. HTTP 429).
	OnRateLimited func(key string, cooldown time.Duration)

	// OnErrorTriggered is invoked when a key records an error and tracks error streaks.
	OnErrorTriggered func(key string, consecutiveErrors int)

	// OnCooldownRecovered is invoked when an expired cooldown key is restored to healthy rotation.
	OnCooldownRecovered func(key string)
}

// Config encapsulates configuration rules for the key rotator pool.
type Config struct {
	// ErrorThreshold specifies consecutive errors before auto-cooldown trips (Default: 3).
	ErrorThreshold int

	// ErrorCooldown is the quarantine duration after reaching ErrorThreshold (Default: 1 minute).
	ErrorCooldown time.Duration

	// DefaultCooldown is the duration applied when marking rate limited without explicit duration (Default: 2 minutes).
	DefaultCooldown time.Duration

	// Hooks provides optional observability callbacks.
	Hooks Hooks
}

// DefaultConfig provides safe, battle-tested defaults.
func DefaultConfig() *Config {
	return &Config{
		ErrorThreshold:  3,
		ErrorCooldown:   1 * time.Minute,
		DefaultCooldown: 2 * time.Minute,
	}
}

// WithErrorThreshold sets consecutive failure count before temporary key suspension.
func WithErrorThreshold(threshold int) Option {
	return func(c *Config) {
		if threshold > 0 {
			c.ErrorThreshold = threshold
		}
	}
}

// WithErrorCooldown sets quarantine duration on error threshold tripping.
func WithErrorCooldown(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.ErrorCooldown = d
		}
	}
}

// WithDefaultCooldown sets default rate-limit backoff period.
func WithDefaultCooldown(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.DefaultCooldown = d
		}
	}
}

// WithHooks attaches zero-dependency telemetry and logging hooks.
func WithHooks(hooks Hooks) Option {
	return func(c *Config) {
		c.Hooks = hooks
	}
}
