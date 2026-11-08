package mimekit

import "errors"

var (
	// ErrEmptyPayload is returned when attempting to sniff an empty byte slice or reader.
	ErrEmptyPayload = errors.New("mimekit: payload is empty")
	// ErrExecutableDetected is returned when uploaded content contains binary executable headers.
	ErrExecutableDetected = errors.New("mimekit: dangerous executable binary detected")
)
