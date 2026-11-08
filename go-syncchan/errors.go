package syncchan

import "errors"

var (
	// ErrNilChannel is returned when attempting to combine nil channels.
	ErrNilChannel = errors.New("syncchan: channel is nil")
	// ErrInvalidCount is returned when requested item count is negative.
	ErrInvalidCount = errors.New("syncchan: count must be non-negative")
)
