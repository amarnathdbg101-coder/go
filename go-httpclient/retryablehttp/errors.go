package retryablehttp

import "errors"

var (
	// ErrMaxRetriesReached is returned when the request has failed after exhausting all retry attempts.
	ErrMaxRetriesReached = errors.New("retryablehttp: max retries reached")
	// ErrNilRequest is returned when an operation is executed on a nil http.Request.
	ErrNilRequest = errors.New("retryablehttp: request is nil")
	// ErrMissingGetBody is returned when retrying a request with a body where GetBody is nil.
	ErrMissingGetBody = errors.New("retryablehttp: request body cannot be rewound; GetBody is nil")
)
