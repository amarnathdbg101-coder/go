package goevents

import "errors"

var (
	// ErrBusClosed indicates operations on a closed EventBus.
	ErrBusClosed = errors.New("events: event bus is closed")

	// ErrQueueFull indicates the asynchronous event buffer is full (backpressure).
	ErrQueueFull = errors.New("events: async event queue is full")

	// ErrInvalidPattern indicates an empty or invalid topic subscription pattern.
	ErrInvalidPattern = errors.New("events: topic pattern cannot be empty")
)
