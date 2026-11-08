package multierror

import "errors"

var (
	// ErrEmptyMultiError is returned when attempting an invalid operation on an empty MultiError.
	ErrEmptyMultiError = errors.New("multierror: no errors accumulated")
)
