package radix

import "errors"

var (
	// ErrKeyNotFound is returned when attempting to delete or access a non-existent key.
	ErrKeyNotFound = errors.New("radix: key not found")
)
