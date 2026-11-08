package gopool

import "errors"

var (
	// ErrPoolClosed indicates acquire attempted on a closed pool.
	ErrPoolClosed = errors.New("gopool: pool is closed")

	// ErrPoolExhausted indicates acquire timed out or pool capacity is fully utilized.
	ErrPoolExhausted = errors.New("gopool: pool capacity exhausted")

	// ErrInvalidConfig indicates pool max capacity <= 0 or nil factory.
	ErrInvalidConfig = errors.New("gopool: invalid pool configuration")
)
