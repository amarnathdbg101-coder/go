package resilience

import "errors"

var (
	// ErrBulkheadFull is returned when all execution slots in the bulkhead are currently occupied.
	ErrBulkheadFull = errors.New("resilience: bulkhead capacity exceeded")

	// ErrCircuitOpen is returned when the circuit breaker is tripped open and rejecting execution.
	ErrCircuitOpen = errors.New("resilience: circuit breaker is open")

	// ErrTimeout is returned when an execution attempt exceeds the configured timeout duration.
	ErrTimeout = errors.New("resilience: execution attempt timed out")

	// ErrMaxAttemptsReached is returned when all retry attempts have been exhausted.
	ErrMaxAttemptsReached = errors.New("resilience: maximum retry attempts exceeded")
)
