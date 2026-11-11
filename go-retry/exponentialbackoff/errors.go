package goexponentialbackoff

import "errors"

var (
	// ErrContextCancelled indicates sleep was interrupted by context cancellation.
	ErrContextCancelled = errors.New("goexponentialbackoff: context cancelled during backoff sleep")
)
