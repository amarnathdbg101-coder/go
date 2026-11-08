package countingbloom

import "errors"

var (
	// ErrInvalidCapacity is returned when expected elements count is non-positive.
	ErrInvalidCapacity = errors.New("countingbloom: capacity must be greater than zero")
	// ErrInvalidFPRate is returned when false positive rate is outside (0, 1).
	ErrInvalidFPRate = errors.New("countingbloom: false positive rate must be between 0 and 1 exclusive")
)
