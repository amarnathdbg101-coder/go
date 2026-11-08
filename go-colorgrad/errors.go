package colorgrad

import "errors"

var (
	// ErrInsufficientStops is returned when fewer than 2 color stops are provided.
	ErrInsufficientStops = errors.New("colorgrad: at least 2 color stops are required")
	// ErrInvalidHex is returned when a hex color string is malformed.
	ErrInvalidHex = errors.New("colorgrad: invalid hex color format")
)
