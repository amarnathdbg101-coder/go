package adaptivepool

import "errors"

var (
	// ErrPoolClosed is returned when submitting tasks to a closed pool.
	ErrPoolClosed = errors.New("adaptivepool: pool is closed")
	// ErrQueueFull is returned when submitting to a full queue in non-blocking mode.
	ErrQueueFull = errors.New("adaptivepool: task queue is full")
	// ErrTimeout is returned when SubmitWithTimeout expires before acquiring a slot.
	ErrTimeout = errors.New("adaptivepool: submit timed out")
	// ErrNilTask is returned when attempting to submit a nil function.
	ErrNilTask = errors.New("adaptivepool: task cannot be nil")
)
