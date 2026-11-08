package jsonpath

import "errors"

var (
	// ErrInvalidPath is returned when the JSONPath syntax is malformed.
	ErrInvalidPath = errors.New("jsonpath: invalid path syntax")

	// ErrNotFound is returned when the specified field does not exist in the JSON structure.
	ErrNotFound = errors.New("jsonpath: key or field not found")

	// ErrIndexOutOfBounds is returned when an array index exceeds bounds or array is empty.
	ErrIndexOutOfBounds = errors.New("jsonpath: array index out of bounds")

	// ErrTypeMismatch is returned when the extracted value cannot be converted to the requested type.
	ErrTypeMismatch = errors.New("jsonpath: type mismatch")

	// ErrInvalidJSON is returned when input data is not valid JSON.
	ErrInvalidJSON = errors.New("jsonpath: invalid json payload")
)
