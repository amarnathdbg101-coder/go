package hpack

import "errors"

var (
	// ErrInvalidIndex is returned when a decoded index is zero or out of bounds.
	ErrInvalidIndex = errors.New("hpack: invalid table index")

	// ErrIntegerOverflow is returned when decoding an integer that exceeds 64-bit bounds.
	ErrIntegerOverflow = errors.New("hpack: integer representation overflow")

	// ErrUnexpectedEOF is returned when the byte stream terminates unexpectedly while parsing.
	ErrUnexpectedEOF = errors.New("hpack: unexpected end of stream")

	// ErrTableSizeExceeded is returned when dynamic table size exceeds configured maximum.
	ErrTableSizeExceeded = errors.New("hpack: dynamic table size exceeded")
)
