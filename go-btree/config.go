package btree

// Config specifies B-Tree parameters.
type Config struct {
	Degree int // Minimum degree t (a node can have between t-1 and 2t-1 keys)
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig provides recommended cache-friendly defaults (Degree 32 -> up to 63 keys per node).
func DefaultConfig() Config {
	return Config{
		Degree: 32,
	}
}

// WithDegree sets the tree branching degree (t >= 2).
func WithDegree(t int) Option {
	return func(c *Config) {
		c.Degree = t
	}
}
