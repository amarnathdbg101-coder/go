package wal

import "errors"

var (
	// ErrCorruptRecord is returned when CRC32 checksum fails during record decoding or replay.
	ErrCorruptRecord = errors.New("wal: data corruption detected, crc32 checksum mismatch")

	// ErrWALClosed is returned when an append or replay is attempted on a closed WAL instance.
	ErrWALClosed = errors.New("wal: write-ahead log is closed")

	// ErrInvalidLSN is returned when requesting replay from an invalid or negative sequence number.
	ErrInvalidLSN = errors.New("wal: invalid log sequence number")

	// ErrPayloadTooLarge is returned when payload exceeds configured maximum record size.
	ErrPayloadTooLarge = errors.New("wal: payload size exceeds maximum segment capacity")
)
