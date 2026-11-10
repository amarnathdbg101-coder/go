package delayqueue

import "errors"

var (
	// ErrJobNotFound is returned when attempting to manipulate an unknown job.
	ErrJobNotFound = errors.New("delayqueue: job not found")

	// ErrQueueClosed is returned when attempting to operate on a stopped queue.
	ErrQueueClosed = errors.New("delayqueue: queue is closed")

	// ErrNoHandler is returned when a job is ready for execution but no handler was registered for its topic.
	ErrNoHandler = errors.New("delayqueue: no handler registered for topic")

	// ErrDuplicateJobID is returned when enqueuing a job with an ID that already exists.
	ErrDuplicateJobID = errors.New("delayqueue: duplicate job id")

	// ErrInvalidDelay is returned when delay duration is negative.
	ErrInvalidDelay = errors.New("delayqueue: delay must be non-negative")
)
