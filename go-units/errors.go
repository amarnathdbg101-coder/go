package gounits

import "errors"

var (
	// ErrInvalidByteString indicates failure to parse byte unit string.
	ErrInvalidByteString = errors.New("gounits: invalid byte string format")

	// ErrInvalidDuration indicates failure to parse human duration string.
	ErrInvalidDuration = errors.New("gounits: invalid duration format")
)
