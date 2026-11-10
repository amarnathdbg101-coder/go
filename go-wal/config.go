package wal

import "time"

// SyncMode determines disk durability behavior.
type SyncMode int

const (
	// SyncAlways calls os.File.Sync (fsync) on every single append for maximum durability.
	SyncAlways SyncMode = iota
	// SyncInterval flushes writes to OS buffer and executes fsync on a periodic background ticker.
	SyncInterval
)

// Config defines configuration parameters for WAL instance.
type Config struct {
	Dir            string
	MaxSegmentSize int64
	SyncMode       SyncMode
	SyncInterval   time.Duration
}

// Option represents a functional option.
type Option func(*Config)

// DefaultConfig provides recommended production defaults.
func DefaultConfig() Config {
	return Config{
		Dir:            "./wal_data",
		MaxSegmentSize: 10 * 1024 * 1024, // 10 MB per segment
		SyncMode:       SyncAlways,
		SyncInterval:   50 * time.Millisecond,
	}
}

// WithDir sets the directory path for storing WAL segment files.
func WithDir(dir string) Option {
	return func(c *Config) {
		c.Dir = dir
	}
}

// WithMaxSegmentSize sets the maximum byte threshold before rolling to a new segment.
func WithMaxSegmentSize(bytes int64) Option {
	return func(c *Config) {
		c.MaxSegmentSize = bytes
	}
}

// WithSyncMode sets the durability sync mode.
func WithSyncMode(mode SyncMode) Option {
	return func(c *Config) {
		c.SyncMode = mode
	}
}

// WithSyncInterval sets the ticker duration for periodic fsyncs in SyncInterval mode.
func WithSyncInterval(d time.Duration) Option {
	return func(c *Config) {
		c.SyncInterval = d
	}
}
