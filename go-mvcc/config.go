package mvcc

// Config defines settings for the MVCC engine.
type Config struct {
	AutoVacuumThreshold int // Number of committed transactions before triggering auto vacuum
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig returns recommended production defaults.
func DefaultConfig() Config {
	return Config{
		AutoVacuumThreshold: 1000,
	}
}

// WithAutoVacuumThreshold sets the transaction interval between automatic vacuum sweeps.
func WithAutoVacuumThreshold(threshold int) Option {
	return func(c *Config) {
		c.AutoVacuumThreshold = threshold
	}
}
