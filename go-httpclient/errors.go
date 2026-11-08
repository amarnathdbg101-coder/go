package gohttpclient

import (
	"errors"
	"fmt"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is active and blocking requests.
	ErrCircuitOpen = errors.New("httpclient: circuit breaker is open (service unavailable)")

	// ErrMaxRetriesExceeded indicates all retry attempts failed.
	ErrMaxRetriesExceeded = errors.New("httpclient: max retries exceeded")

	// ErrInvalidStatus indicates an unexpected HTTP response status code.
	ErrInvalidStatus = errors.New("httpclient: unexpected status code")
)

// StatusError represents an error response containing the HTTP status code and body snippet.
type StatusError struct {
	StatusCode int
	Body       string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("httpclient: status %d: %s", e.StatusCode, e.Body)
}

func (e *StatusError) Unwrap() error {
	return ErrInvalidStatus
}
