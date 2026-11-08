package gobloom

import "errors"

var (
	// ErrInvalidParams indicates invalid expectedItems or falsePositiveRate values.
	ErrInvalidParams = errors.New("bloom: expected items must be > 0 and false positive rate between 0 and 1")

	// ErrCorruptData indicates corrupted serialized bloom filter bytes.
	ErrCorruptData = errors.New("bloom: corrupted filter byte representation")
)
