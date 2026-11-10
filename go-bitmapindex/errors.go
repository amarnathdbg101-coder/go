package bitmapindex

import "errors"

var (
	// ErrBitWidthExceeded is returned when a value exceeds the maximum bit-width capacity of the index.
	ErrBitWidthExceeded = errors.New("bitmapindex: value exceeds configured bit-width capacity")

	// ErrRowNotFound is returned when querying a row ID that has no value set.
	ErrRowNotFound = errors.New("bitmapindex: row id has no indexed value")

	// ErrInvalidRange is returned when minVal > maxVal in range queries.
	ErrInvalidRange = errors.New("bitmapindex: min value cannot be greater than max value")
)
