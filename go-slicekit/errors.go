package slicekit

import "errors"

var (
	// ErrInvalidChunkSize is returned when chunk size is <= 0.
	ErrInvalidChunkSize = errors.New("slicekit: chunk size must be greater than 0")

	// ErrEmptySlice is returned when operating on an empty slice where disallowed.
	ErrEmptySlice = errors.New("slicekit: slice is empty")
)
