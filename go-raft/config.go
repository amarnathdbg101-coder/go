package raft

import "time"

// Config defines timing and identity settings for a Raft node.
type Config struct {
	ID                string
	Peers             []string
	HeartbeatInterval time.Duration
	ElectionTimeout   time.Duration
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig provides recommended production defaults.
func DefaultConfig(id string, peers []string) Config {
	return Config{
		ID:                id,
		Peers:             peers,
		HeartbeatInterval: 50 * time.Millisecond,
		ElectionTimeout:   200 * time.Millisecond,
	}
}

// WithHeartbeatInterval sets the leader heartbeat cadence.
func WithHeartbeatInterval(d time.Duration) Option {
	return func(c *Config) {
		c.HeartbeatInterval = d
	}
}

// WithElectionTimeout sets the base election timeout before initiating a vote.
func WithElectionTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.ElectionTimeout = d
	}
}
