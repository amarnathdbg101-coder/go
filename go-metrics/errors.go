package gometrics

import "errors"

var (
	// ErrDuplicateMetric occurs when registering a metric with an already registered name.
	ErrDuplicateMetric = errors.New("metrics: metric already registered")

	// ErrInvalidMetricName occurs when a metric name contains invalid characters.
	ErrInvalidMetricName = errors.New("metrics: invalid metric name (must match [a-zA-Z_:][a-zA-Z0-9_:]*)")

	// ErrMetricNotFound occurs when looking up a metric that does not exist.
	ErrMetricNotFound = errors.New("metrics: metric not found")
)
