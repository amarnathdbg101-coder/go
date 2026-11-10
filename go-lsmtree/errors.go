package lsmtree

import "errors"

var (
	// ErrKeyNotFound is returned when a requested key does not exist or has been deleted.
	ErrKeyNotFound = errors.New("lsmtree: key not found")

	// ErrDBClosed is returned when an operation is executed on a closed LSM database.
	ErrDBClosed = errors.New("lsmtree: database is closed")

	// ErrEmptyKey is returned when passing a zero-length key.
	ErrEmptyKey = errors.New("lsmtree: key cannot be empty")

	// ErrCorruptWAL is returned when an entry in the write-ahead log fails checksum validation.
	ErrCorruptWAL = errors.New("lsmtree: corrupt write-ahead log entry")
)
