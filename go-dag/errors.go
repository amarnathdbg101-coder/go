package dag

import "errors"

var (
	// ErrCycleDetected is returned when the graph contains circular dependencies.
	ErrCycleDetected = errors.New("dag: cycle detected in graph")
	// ErrNodeNotFound is returned when an edge references a non-existent node.
	ErrNodeNotFound = errors.New("dag: node not found")
	// ErrDuplicateNode is returned when adding a node with an ID that already exists.
	ErrDuplicateNode = errors.New("dag: duplicate node ID")
	// ErrNilTask is returned when attempting to register a node with a nil TaskFunc.
	ErrNilTask = errors.New("dag: task function is nil")
)
