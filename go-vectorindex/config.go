package vectorindex

// Metric defines the distance / similarity calculation method.
type Metric int

const (
	// MetricCosine calculates Cosine Similarity in [-1.0, 1.0] (higher is more similar).
	MetricCosine Metric = iota
	// MetricEuclidean calculates L2 distance in [0, inf) (lower is more similar).
	MetricEuclidean
	// MetricDotProduct calculates inner product (higher is more similar for normalized vectors).
	MetricDotProduct
)

// Config defines settings for the Vector Index.
type Config struct {
	Dimensions int    // Number of vector dimensions (e.g. 128, 384, 768, 1536)
	Metric     Metric // Distance metric
}

// Option configures functional options.
type Option func(*Config)

// DefaultConfig provides standard recommended defaults.
func DefaultConfig(dimensions int) Config {
	return Config{
		Dimensions: dimensions,
		Metric:     MetricCosine,
	}
}

// WithMetric sets the distance metric.
func WithMetric(m Metric) Option {
	return func(c *Config) {
		c.Metric = m
	}
}
