package gochecksum

import "errors"

var (
	// ErrUnsupportedAlgorithm indicates unknown checksum algorithm.
	ErrUnsupportedAlgorithm = errors.New("gochecksum: unsupported checksum algorithm")

	// ErrChecksumMismatch indicates calculated checksum does not match expected value.
	ErrChecksumMismatch = errors.New("gochecksum: checksum mismatch")
)
