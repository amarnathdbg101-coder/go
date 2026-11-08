package dedup

import "errors"

var (
	// ErrClosed is returned when an operation is performed on a closed deduplicator.
	ErrClosed = errors.New("dedup: deduplicator is closed")
)
