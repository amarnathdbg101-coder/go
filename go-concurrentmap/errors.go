package goconcurrentmap

import "errors"

var (
	// ErrKeyNotFound indicates key does not exist in the concurrent map.
	ErrKeyNotFound = errors.New("goconcurrentmap: key not found")
)
