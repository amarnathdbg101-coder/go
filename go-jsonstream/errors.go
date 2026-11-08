package jsonstream

import "errors"

var (
	// ErrExpectedArray is returned when the JSON stream does not begin with '['.
	ErrExpectedArray = errors.New("jsonstream: expected '[' at beginning of array stream")

	// ErrMalformedJSON is returned when stream encounters invalid or truncated JSON.
	ErrMalformedJSON = errors.New("jsonstream: malformed or truncated json")

	// ErrWriterClosed is returned when attempting to write to a closed streaming writer.
	ErrWriterClosed = errors.New("jsonstream: writer is already closed")
)
