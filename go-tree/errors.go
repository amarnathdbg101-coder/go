package gotree

import "errors"

var (
	// ErrNodeNotFound indicates the requested node was not found in the tree.
	ErrNodeNotFound = errors.New("gotree: node not found")

	// ErrDuplicateNode indicates attempting to insert a node with an existing identifier.
	ErrDuplicateNode = errors.New("gotree: duplicate node identifier")

	// ErrParentNotFound indicates the specified parent node does not exist.
	ErrParentNotFound = errors.New("gotree: parent node not found")

	// ErrCycleDetected indicates linking a node would create a circular loop in the hierarchy.
	ErrCycleDetected = errors.New("gotree: cycle detected in hierarchy")
)
