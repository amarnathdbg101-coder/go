package bytespool

import "errors"

var (
	// ErrNilBuffer is returned when attempting to put or operate on a nil buffer.
	ErrNilBuffer = errors.New("bytespool: buffer is nil")
)
