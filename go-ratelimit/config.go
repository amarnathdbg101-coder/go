package ratelimit

import (
	"errors"
	"time"
)

// Config holds runtime configuration options for go-ratelimit.
type Config struct {
	Capacity        int
	Window          time.Duration
	KeyFunc         KeyFunc
	OnLimitExceeded ErrorHandler
}

// DefaultConfig provides sensible defaults (100 requests / minute by IP).
func DefaultConfig() Config {
	return Config{
		Capacity:        100,
		Window:          1 * time.Minute,
		KeyFunc:         KeyByIP,
		OnLimitExceeded: DefaultErrorHandler,
	}
}

// Validate checks configuration parameters.
func (c *Config) Validate() error {
	if c.Capacity <= 0 {
		return errors.New("ratelimit: capacity must be greater than zero")
	}
	if c.Window <= 0 {
		return errors.New("ratelimit: window duration must be greater than zero")
	}
	if c.KeyFunc == nil {
		c.KeyFunc = KeyByIP
	}
	if c.OnLimitExceeded == nil {
		c.OnLimitExceeded = DefaultErrorHandler
	}
	return nil
}
