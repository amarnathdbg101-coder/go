package sstable

// Config defines the storage parameters for building an SSTable.
type Config struct {
	BlockSize        int     // Maximum data block size before flushing (default 4KB)
	BloomBitsPerKey  int     // Number of bloom filter bits per key (default 10 bits -> ~1% false positive)
}

// Option represents a functional option.
type Option func(*Config)

// DefaultConfig provides standard LSM database defaults.
func DefaultConfig() Config {
	return Config{
		BlockSize:       4096,
		BloomBitsPerKey: 10,
	}
}

// WithBlockSize sets the uncompressed data block target size in bytes.
func WithBlockSize(bytes int) Option {
	return func(c *Config) {
		c.BlockSize = bytes
	}
}

// WithBloomBitsPerKey sets the bloom filter resolution (bits per key).
func WithBloomBitsPerKey(bits int) Option {
	return func(c *Config) {
		c.BloomBitsPerKey = bits
	}
}
