package taskqueue

import "errors"

// Common domain errors for go-taskqueue.
var (
	// ErrQueueFull is returned when non-blocking enqueue fails because buffer capacity is reached.
	ErrQueueFull = errors.New("taskqueue: task buffer is full")

	// ErrQueueClosed is returned when attempting to enqueue onto a stopped queue.
	ErrQueueClosed = errors.New("taskqueue: queue has been shut down")

	// ErrHandlerNotFound is returned when no processor function has been registered for a task type.
	ErrHandlerNotFound = errors.New("taskqueue: no handler registered for task type")

	// ErrTaskTimeout is returned when a task exceeds its configured execution timeout.
	ErrTaskTimeout = errors.New("taskqueue: task execution timed out")

	// ErrMaxRetriesExceeded is returned when all retry attempts for a task have failed.
	ErrMaxRetriesExceeded = errors.New("taskqueue: maximum retry attempts exceeded")
)
