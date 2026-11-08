package cachekit

import "time"

// Config holds runtime configuration options for go-cachekit Engine.
type Config struct {
	DefaultTTL time.Duration
}

// DefaultConfig provides sensible defaults (10 minutes TTL).
func DefaultConfig() Config {
	return Config{
		DefaultTTL: 10 * time.Minute,
	}
}

// Validate checks configuration.
func (c *Config) Validate() error {
	if c.DefaultTTL <= 0 {
		c.DefaultTTL = 10 * time.Minute
	}
	return nil
}
