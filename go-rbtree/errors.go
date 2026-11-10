package rbtree

import "errors"

var (
	// ErrKeyNotFound is returned when attempting to delete a non-existent key.
	ErrKeyNotFound = errors.New("rbtree: key not found")

	// ErrEmptyTree is returned when requesting Min/Max/Floor/Ceiling on an empty tree.
	ErrEmptyTree = errors.New("rbtree: tree is empty")
)
