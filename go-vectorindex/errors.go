package vectorindex

import "errors"

var (
	// ErrDimensionMismatch is returned when query vector dimensions do not match the index dimensions.
	ErrDimensionMismatch = errors.New("vectorindex: vector dimension does not match index dimension")

	// ErrZeroVector is returned when calculating cosine similarity on a vector with zero magnitude.
	ErrZeroVector = errors.New("vectorindex: cannot normalize or calculate cosine on zero-magnitude vector")

	// ErrInvalidK is returned when requesting K <= 0 neighbors.
	ErrInvalidK = errors.New("vectorindex: top-k parameter must be at least 1")

	// ErrItemNotFound is returned when attempting to delete a non-existent vector ID.
	ErrItemNotFound = errors.New("vectorindex: item ID not found in index")
)
