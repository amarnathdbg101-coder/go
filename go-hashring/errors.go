package hashring

import "errors"

var (
	// ErrEmptyRing is returned when querying a hash ring that contains no physical nodes.
	ErrEmptyRing = errors.New("hashring: hash ring is empty")

	// ErrNodeNotFound is returned when attempting to remove a node that does not exist in the ring.
	ErrNodeNotFound = errors.New("hashring: node not found in ring")

	// ErrInvalidCount is returned when requesting <= 0 nodes.
	ErrInvalidCount = errors.New("hashring: requested count must be greater than 0")
)
