package dataloader

import "errors"

var (
	// ErrBatchTooSmall is returned when a batch has no keys.
	ErrBatchTooSmall = errors.New("dataloader: batch has no keys")

	// ErrResultLengthMismatch is returned when batch function returns fewer or more results than keys requested.
	ErrResultLengthMismatch = errors.New("dataloader: batch function returned mismatched result count")

	// ErrLoaderClosed is returned when attempting to load on a closed loader.
	ErrLoaderClosed = errors.New("dataloader: loader is closed")

	// ErrKeyNotFound is returned when a batch function omits a requested key in map-based dispatch.
	ErrKeyNotFound = errors.New("dataloader: key not found in batch results")
)
