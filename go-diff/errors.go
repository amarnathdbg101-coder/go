package godiff

import "errors"

var (
	// ErrTypeMismatch indicates comparing incompatible data types.
	ErrTypeMismatch = errors.New("godiff: cannot compare different types")

	// ErrInvalidInput indicates comparison on unsupported primitives.
	ErrInvalidInput = errors.New("godiff: invalid input for diff calculation")
)
