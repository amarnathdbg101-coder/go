package lfu

import "errors"

var (
	// ErrInvalidCapacity is returned when creating an LFU cache with capacity <= 0.
	ErrInvalidCapacity = errors.New("lfu: capacity must be greater than 0")
)
