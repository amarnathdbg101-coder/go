package latch

import "errors"

var (
	// ErrTimeout is returned when waiting exceeds the allowed duration.
	ErrTimeout = errors.New("latch: wait timeout exceeded")

	// ErrBrokenBarrier is returned when a barrier is broken by another thread or reset.
	ErrBrokenBarrier = errors.New("latch: cyclic barrier is broken")

	// ErrInvalidCount is returned when creating a latch or barrier with count <= 0.
	ErrInvalidCount = errors.New("latch: count/parties must be greater than 0")
)
