package simhash

import "errors"

var (
	// ErrEmptyInput is returned when hashing or indexing an empty feature set or string.
	ErrEmptyInput = errors.New("simhash: input text or feature set cannot be empty")

	// ErrInvalidDistance is returned when a query requests a negative or out-of-bounds Hamming distance.
	ErrInvalidDistance = errors.New("simhash: hamming distance must be between 0 and 64")

	// ErrItemNotFound is returned when an ID is not found in the SimHash index.
	ErrItemNotFound = errors.New("simhash: item not found in index")
)
