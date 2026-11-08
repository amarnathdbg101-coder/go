package gosafecast

import "errors"

var (
	// ErrOverflow indicates the value is too large for the destination type.
	ErrOverflow = errors.New("gosafecast: numeric overflow")

	// ErrUnderflow indicates the value is too small (e.g. negative into unsigned) for destination type.
	ErrUnderflow = errors.New("gosafecast: numeric underflow")

	// ErrNonInteger indicates a float value has a fractional part when integer is required.
	ErrNonInteger = errors.New("gosafecast: fractional float cannot be safely converted to integer")
)
