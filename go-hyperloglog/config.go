package hyperloglog

// Config specifies the precision of the HyperLogLog sketch.
type Config struct {
	Precision uint8 // Precision p in [4, 18], defines m = 2^p registers
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig provides recommended production defaults (p=14 -> 16,384 registers, ~0.81% error).
func DefaultConfig() Config {
	return Config{
		Precision: 14,
	}
}

// WithPrecision sets the precision bit size (4 <= p <= 18).
func WithPrecision(p uint8) Option {
	return func(c *Config) {
		c.Precision = p
	}
}
