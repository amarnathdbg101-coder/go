package trafficshaper

import "errors"

var (
	// ErrQueueFull is returned when the shaping queue has reached its maximum buffer limit.
	ErrQueueFull = errors.New("trafficshaper: queue buffer is full")

	// ErrShaperClosed is returned when an operation is attempted on a stopped traffic shaper.
	ErrShaperClosed = errors.New("trafficshaper: shaper is stopped or closed")

	// ErrContextTimeout is returned when a caller's context deadline expires while queued.
	ErrContextTimeout = errors.New("trafficshaper: context deadline exceeded while waiting in queue")

	// ErrInvalidConfig is returned when configuration values are non-positive or invalid.
	ErrInvalidConfig = errors.New("trafficshaper: invalid configuration parameters")
)
