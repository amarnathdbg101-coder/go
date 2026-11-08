package godebounce

import "errors"

var (
	// ErrInvalidDuration indicates debounce/throttle wait duration is non-positive.
	ErrInvalidDuration = errors.New("godebounce: wait duration must be greater than zero")

	// ErrActionNil indicates function callback is nil.
	ErrActionNil = errors.New("godebounce: action cannot be nil")
)
