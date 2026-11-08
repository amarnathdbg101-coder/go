package gosecureid

import "errors"

var (
	// ErrInvalidLength indicates length <= 0.
	ErrInvalidLength = errors.New("gosecureid: length must be greater than zero")

	// ErrEmptyAlphabet indicates alphabet must contain at least 2 characters.
	ErrEmptyAlphabet = errors.New("gosecureid: alphabet must contain at least 2 characters")
)
