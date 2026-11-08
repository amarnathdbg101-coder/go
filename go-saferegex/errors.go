package saferegex

import "errors"

var (
	// ErrTimeout is returned when regular expression execution exceeds the configured deadline.
	ErrTimeout = errors.New("saferegex: execution timed out (potential ReDoS)")
	// ErrInvalidPattern is returned when regex pattern compilation fails.
	ErrInvalidPattern = errors.New("saferegex: invalid regular expression pattern")
)
