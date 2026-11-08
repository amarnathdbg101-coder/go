package set

import "errors"

var (
	// ErrNilSet is returned when attempting an operation on a nil Set.
	ErrNilSet = errors.New("set: set operand is nil")
)
