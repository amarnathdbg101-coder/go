package crdt

import "errors"

var (
	// ErrNilNode is returned when attempting to merge with a nil CRDT state.
	ErrNilNode = errors.New("crdt: cannot merge with nil CRDT instance")

	// ErrInvalidNodeID is returned when initializing a CRDT with an empty node identifier.
	ErrInvalidNodeID = errors.New("crdt: node identifier cannot be empty")

	// ErrElementNotFound is returned when attempting to remove an element that does not exist in an ORSet.
	ErrElementNotFound = errors.New("crdt: element not found in observed set")
)
