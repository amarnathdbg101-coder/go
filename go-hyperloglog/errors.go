package hyperloglog

import "errors"

var (
	// ErrInvalidPrecision is returned when precision p is outside the valid range [4, 18].
	ErrInvalidPrecision = errors.New("hyperloglog: precision p must be between 4 and 18")

	// ErrPrecisionMismatch is returned when attempting to merge two sketches with different precision parameters.
	ErrPrecisionMismatch = errors.New("hyperloglog: cannot merge sketches with different precision values")

	// ErrNilSketch is returned when passing a nil sketch to Merge.
	ErrNilSketch = errors.New("hyperloglog: sketch cannot be nil")
)
