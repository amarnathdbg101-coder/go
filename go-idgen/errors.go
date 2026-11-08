package goidgen

import "errors"

var (
	// ErrInvalidNodeID occurs when node ID exceeds 10-bit limit (0 to 1023).
	ErrInvalidNodeID = errors.New("idgen: node id must be between 0 and 1023")

	// ErrClockBackwards occurs when the system clock drifts backwards beyond safe recovery threshold.
	ErrClockBackwards = errors.New("idgen: clock drifted backwards, refusing to generate duplicate id")

	// ErrInvalidBase62 occurs when decoding an invalid Base62 string.
	ErrInvalidBase62 = errors.New("idgen: invalid base62 character")
)
