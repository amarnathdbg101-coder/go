package gobitmap

import "errors"

var (
	// ErrOutOfBounds indicates the requested bit index is negative.
	ErrOutOfBounds = errors.New("gobitmap: bit index cannot be negative")

	// ErrSizeMismatch indicates binary operations attempted on bitmaps of incompatible capacities.
	ErrSizeMismatch = errors.New("gobitmap: bitmap size mismatch")
)
