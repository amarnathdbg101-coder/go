package envconfig

import "errors"

var (
	// ErrNotPointerToStruct is returned when the target config passed to Load is not a pointer to a struct.
	ErrNotPointerToStruct = errors.New("envconfig: target must be a non-nil pointer to a struct")

	// ErrMissingRequired is returned when a required environment variable is missing or empty with no default.
	ErrMissingRequired = errors.New("envconfig: missing required environment variable")

	// ErrUnsupportedType is returned when a struct field type is not supported for unmarshaling.
	ErrUnsupportedType = errors.New("envconfig: unsupported field type")

	// ErrInvalidValue is returned when an environment variable string cannot be parsed into the target type.
	ErrInvalidValue = errors.New("envconfig: cannot parse value into field type")
)
