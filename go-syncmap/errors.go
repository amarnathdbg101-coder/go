package syncmap

import "errors"

var (
	// ErrKeyNotFound is returned when an expected key is missing in the map.
	ErrKeyNotFound = errors.New("syncmap: key not found")
)
