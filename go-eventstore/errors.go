package eventstore

import "errors"

var (
	// ErrWrongExpectedVersion is returned when a stream append violates optimistic concurrency control.
	ErrWrongExpectedVersion = errors.New("eventstore: wrong expected stream revision, concurrent modification detected")

	// ErrStreamNotFound is returned when reading from a non-existent event stream.
	ErrStreamNotFound = errors.New("eventstore: event stream not found")

	// ErrEmptyEvents is returned when attempting to append zero events.
	ErrEmptyEvents = errors.New("eventstore: cannot append empty event slice")

	// ErrSnapshotNotFound is returned when no snapshot exists for the requested stream.
	ErrSnapshotNotFound = errors.New("eventstore: snapshot not found for stream")
)
