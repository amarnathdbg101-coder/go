package mvcc

import "errors"

var (
	// ErrKeyNotFound is returned when no visible version of the requested key exists for the transaction snapshot.
	ErrKeyNotFound = errors.New("mvcc: key not found in transaction snapshot")

	// ErrWriteConflict is returned when attempting to write to a key that has been concurrently modified (First-Committer-Wins).
	ErrWriteConflict = errors.New("mvcc: write-write conflict detected, concurrent transaction committed newer version")

	// ErrTxClosed is returned when attempting operations on a committed or aborted transaction.
	ErrTxClosed = errors.New("mvcc: transaction is already committed or rolled back")

	// ErrEmptyKey is returned when passing an empty key.
	ErrEmptyKey = errors.New("mvcc: key cannot be empty")
)
