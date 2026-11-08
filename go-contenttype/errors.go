package contenttype

import "errors"

var (
	// ErrEmptyHeader is returned when the content header is empty.
	ErrEmptyHeader = errors.New("contenttype: header string is empty")

	// ErrInvalidMediaType is returned when media type format is malformed (lacks type/subtype slash).
	ErrInvalidMediaType = errors.New("contenttype: invalid media type format")
)
