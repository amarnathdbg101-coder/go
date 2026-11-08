package validator

// Config holds runtime options for the Validator.
type Config struct {
	// StopOnFirstError halts validation immediately upon encountering the first failure.
	// Default: false (collects all failed fields so frontend can display all form errors at once).
	StopOnFirstError bool

	// TrimStrings automatically trims whitespace before applying rules.
	// Default: true.
	TrimStrings bool
}

// Option customizes Config behavior.
type Option func(*Config)

// DefaultConfig returns standard production validation settings.
func DefaultConfig() Config {
	return Config{
		StopOnFirstError: false,
		TrimStrings:      true,
	}
}

// WithStopOnFirstError configures validator to exit on the first failed check.
func WithStopOnFirstError(stop bool) Option {
	return func(c *Config) {
		c.StopOnFirstError = stop
	}
}

// WithTrimStrings toggles automatic whitespace trimming.
func WithTrimStrings(trim bool) Option {
	return func(c *Config) {
		c.TrimStrings = trim
	}
}
