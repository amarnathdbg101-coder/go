package cuckoofilter

// Config defines the parameters for the Cuckoo Filter.
type Config struct {
	Capacity uint // Maximum number of items the filter should store
	MaxKicks uint // Maximum cuckoo kicks allowed before declaring filter full (default 500)
}

// Option represents a functional option.
type Option func(*Config)

// DefaultConfig provides standard recommended defaults.
func DefaultConfig() Config {
	return Config{
		Capacity: 10000,
		MaxKicks: 500,
	}
}

// WithCapacity sets the target item capacity.
func WithCapacity(cap uint) Option {
	return func(c *Config) {
		c.Capacity = cap
	}
}

// WithMaxKicks sets the maximum displacement steps.
func WithMaxKicks(kicks uint) Option {
	return func(c *Config) {
		c.MaxKicks = kicks
	}
}
