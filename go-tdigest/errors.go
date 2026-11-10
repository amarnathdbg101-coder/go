package tdigest

import "errors"

var (
	// ErrInvalidCompression is returned when compression delta parameter is non-positive.
	ErrInvalidCompression = errors.New("tdigest: compression parameter must be greater than 0")

	// ErrInvalidQuantile is returned when the requested quantile q is outside [0.0, 1.0].
	ErrInvalidQuantile = errors.New("tdigest: quantile q must be between 0.0 and 1.0")

	// ErrEmptyDigest is returned when estimating a quantile on an empty digest.
	ErrEmptyDigest = errors.New("tdigest: digest contains no samples")
)
