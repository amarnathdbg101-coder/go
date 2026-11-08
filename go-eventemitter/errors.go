package eventemitter

import "errors"

var (
	// ErrNilListener is returned when registering a nil listener callback.
	ErrNilListener = errors.New("eventemitter: listener function cannot be nil")
	// ErrEmptyEvent is returned when attempting to register or emit a blank event name.
	ErrEmptyEvent = errors.New("eventemitter: event name cannot be empty")
)
