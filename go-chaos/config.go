package chaos

// Config defines the global operational flags for the chaos engine.
type Config struct {
	// Enabled is the master killswitch for all chaos injection. Default is true.
	Enabled bool

	// BypassHeader specifies an HTTP header that completely bypasses fault injection (e.g. for health checks).
	// Default is "X-Bypass-Chaos".
	BypassHeader string

	// BypassToken is an optional secret token required in the BypassHeader to allow bypass.
	BypassToken string
}

// Option is a functional option for configuring the Chaos engine.
type Option func(*Config)

// DefaultConfig returns production-safe defaults for Chaos.
func DefaultConfig() Config {
	return Config{
		Enabled:      true,
		BypassHeader: "X-Bypass-Chaos",
	}
}

// WithEnabled toggles the master chaos engine switch.
func WithEnabled(enabled bool) Option {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

// WithBypassHeader configures a custom bypass header.
func WithBypassHeader(header, token string) Option {
	return func(c *Config) {
		c.BypassHeader = header
		c.BypassToken = token
	}
}
