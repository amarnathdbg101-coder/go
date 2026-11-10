package heavyhitters

import "errors"

var (
	// ErrInvalidCapacity is returned when the capacity parameter is non-positive.
	ErrInvalidCapacity = errors.New("heavyhitters: capacity must be greater than 0")

	// ErrInvalidK is returned when requesting a top-k count less than 1.
	ErrInvalidK = errors.New("heavyhitters: top-k count must be at least 1")
)
