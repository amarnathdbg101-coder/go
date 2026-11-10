package sstable

import "errors"

var (
	// ErrKeyNotFound is returned when a requested key does not exist in the SSTable.
	ErrKeyNotFound = errors.New("sstable: key not found")

	// ErrUnsortedKeys is returned when attempting to append a key smaller than or equal to the previous key.
	ErrUnsortedKeys = errors.New("sstable: keys must be inserted in strictly ascending lexicographical order")

	// ErrCorruptFooter is returned when the SSTable file footer magic or offsets are invalid.
	ErrCorruptFooter = errors.New("sstable: invalid or corrupt sstable footer")

	// ErrWriterClosed is returned when an operation is performed on a closed writer.
	ErrWriterClosed = errors.New("sstable: writer is already finished or closed")
)
