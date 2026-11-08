package gobatcher

import "errors"

var (
	// ErrBatcherStopped indicates attempting to push items into a stopped batcher.
	ErrBatcherStopped = errors.New("gobatcher: batcher is stopped")

	// ErrQueueFull indicates the input queue buffer has exceeded maximum capacity.
	ErrQueueFull = errors.New("gobatcher: queue is full (backpressure)")

	// ErrInvalidConfig indicates batch size, buffer, or timeout configuration is invalid.
	ErrInvalidConfig = errors.New("gobatcher: invalid batcher configuration")
)
