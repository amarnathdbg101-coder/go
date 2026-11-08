package gotemplatekit

import "errors"

var (
	// ErrMissingVariable indicates a template tag has no corresponding data in strict mode.
	ErrMissingVariable = errors.New("gotemplatekit: missing variable in strict mode")

	// ErrUnknownFilter indicates an unregistered filter name in pipe expression.
	ErrUnknownFilter = errors.New("gotemplatekit: unknown filter function")
)
