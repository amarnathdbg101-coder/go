package gokvstore

import "errors"

var (
	// ErrKeyNotFound indicates the requested key does not exist or has expired.
	ErrKeyNotFound = errors.New("gokvstore: key not found")

	// ErrNotAnInteger indicates increment was attempted on a non-integer value.
	ErrNotAnInteger = errors.New("gokvstore: value is not an integer")

	// ErrStoreClosed indicates operations were attempted on a closed store.
	ErrStoreClosed = errors.New("gokvstore: store is closed")
)
