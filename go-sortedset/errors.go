package sortedset

import "errors"

var (
	// ErrMemberNotFound is returned when operating on a non-existent member.
	ErrMemberNotFound = errors.New("sortedset: member not found")
	// ErrEmptySet is returned when popping from an empty sorted set.
	ErrEmptySet = errors.New("sortedset: set is empty")
)
