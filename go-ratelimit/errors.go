package ratelimit

import "errors"

// Standard domain errors returned by go-ratelimit.
var (
	ErrLimitExceeded = errors.New("rate limit exceeded")
	ErrNilStore      = errors.New("ratelimit: store must not be nil")
	ErrInvalidConfig = errors.New("ratelimit: invalid configuration")
)
