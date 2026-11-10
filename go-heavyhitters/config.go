package heavyhitters

// Config specifies parameters for the Space-Saving Heavy Hitters tracker.
type Config struct {
	Capacity int // Number of internal tracking slots (M = 1 / epsilon)
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig provides recommended production defaults (Capacity 1000 -> 0.1% error bound).
func DefaultConfig() Config {
	return Config{
		Capacity: 1000,
	}
}

// WithCapacity sets the internal tracking capacity.
func WithCapacity(cap int) Option {
	return func(c *Config) {
		c.Capacity = cap
	}
}
