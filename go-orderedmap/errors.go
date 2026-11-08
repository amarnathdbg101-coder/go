package orderedmap

import "errors"

var (
	// ErrKeyNotFound is returned when a requested key does not exist.
	ErrKeyNotFound = errors.New("orderedmap: key not found")
	// ErrNilMap is returned when an operation is performed on an uninitialized map.
	ErrNilMap = errors.New("orderedmap: map is nil")
	// ErrInvalidJSON is returned when JSON decoding fails or structure is malformed.
	ErrInvalidJSON = errors.New("orderedmap: invalid json payload")
)
