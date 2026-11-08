package pathkit

import "errors"

var (
	// ErrPathTraversal is returned when a path resolves outside the root base directory.
	ErrPathTraversal = errors.New("pathkit: path traversal detected")
	// ErrEmptyPath is returned when a provided path is blank.
	ErrEmptyPath = errors.New("pathkit: path is empty")
	// ErrMaxDepthExceeded is returned when directory walking reaches maximum depth.
	ErrMaxDepthExceeded = errors.New("pathkit: max directory depth exceeded")
)
