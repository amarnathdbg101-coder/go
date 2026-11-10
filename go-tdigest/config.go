package tdigest

// Config defines settings for the t-Digest online streaming quantile estimator.
type Config struct {
	Compression float64 // Compression delta (typically 100 to 500, default 100)
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig provides recommended production defaults.
func DefaultConfig() Config {
	return Config{
		Compression: 100.0,
	}
}

// WithCompression sets the compression parameter delta (higher = more centroids, higher accuracy).
func WithCompression(delta float64) Option {
	return func(c *Config) {
		c.Compression = delta
	}
}
