package urlquery

import "errors"

var (
	// ErrNotPointerToStruct is returned when the target object passed to Decode is not a non-nil pointer to a struct.
	ErrNotPointerToStruct = errors.New("urlquery: target must be a non-nil pointer to a struct")

	// ErrNotStruct is returned when the source object passed to Encode is not a struct or pointer to struct.
	ErrNotStruct = errors.New("urlquery: source must be a struct or pointer to struct")

	// ErrUnsupportedType is returned when a field type is not supported for encoding/decoding.
	ErrUnsupportedType = errors.New("urlquery: unsupported field type")

	// ErrInvalidValue is returned when a query parameter value cannot be parsed into the target field.
	ErrInvalidValue = errors.New("urlquery: cannot parse value into field type")
)
