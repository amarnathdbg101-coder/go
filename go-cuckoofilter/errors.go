package cuckoofilter

import "errors"

var (
	// ErrFilterFull is returned when max cuckoo kicks are exceeded during insertion.
	ErrFilterFull = errors.New("cuckoofilter: filter is full, relocation kicks exceeded")

	// ErrInvalidCapacity is returned when the configured capacity is non-positive.
	ErrInvalidCapacity = errors.New("cuckoofilter: capacity must be greater than 0")
)
