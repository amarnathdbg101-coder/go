package lsmtree

// Config defines the storage parameters for the LSM-Tree engine.
type Config struct {
	Dir              string // Directory where WAL and SSTable files are stored
	MemTableCapacity int    // Maximum bytes in active MemTable before triggering flush
	SyncWAL          bool   // Whether to fsync WAL on every write
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig provides standard recommended production defaults.
func DefaultConfig() Config {
	return Config{
		Dir:              "./lsm_data",
		MemTableCapacity: 4 * 1024 * 1024, // 4 MB
		SyncWAL:          true,
	}
}

// WithDir sets the directory path for database files.
func WithDir(dir string) Option {
	return func(c *Config) {
		c.Dir = dir
	}
}

// WithMemTableCapacity sets the memory threshold before flushing to SSTable.
func WithMemTableCapacity(bytes int) Option {
	return func(c *Config) {
		c.MemTableCapacity = bytes
	}
}

// WithSyncWAL enables or disables synchronous fsync on WAL append.
func WithSyncWAL(sync bool) Option {
	return func(c *Config) {
		c.SyncWAL = sync
	}
}
