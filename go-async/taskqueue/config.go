package taskqueue

import (
	"errors"
	"time"
)

type Config struct {
	Workers            int
	BufferSize         int
	DefaultMaxRetries  int
	InitialRetryDelay  time.Duration
	MaxRetryDelay      time.Duration
}

func DefaultConfig() Config {
	return Config{
		Workers:           10,
		BufferSize:        1000,
		DefaultMaxRetries: 3,
		InitialRetryDelay: 1 * time.Second,
		MaxRetryDelay:     30 * time.Second,
	}
}

func (c *Config) Validate() error {
	if c.Workers <= 0 {
		return errors.New("taskqueue: workers must be greater than zero")
	}
	if c.BufferSize <= 0 {
		c.BufferSize = 1000
	}
	if c.DefaultMaxRetries < 0 {
		c.DefaultMaxRetries = 3
	}
	if c.InitialRetryDelay <= 0 {
		c.InitialRetryDelay = 1 * time.Second
	}
	if c.MaxRetryDelay <= 0 {
		c.MaxRetryDelay = 30 * time.Second
	}
	return nil
}
