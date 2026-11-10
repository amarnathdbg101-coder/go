package btree

import "errors"

var (
	// ErrInvalidDegree is returned when B-Tree degree is less than 2.
	ErrInvalidDegree = errors.New("btree: degree must be at least 2")

	// ErrKeyNotFound is returned when attempting to delete a non-existent key.
	ErrKeyNotFound = errors.New("btree: key not found")
)
