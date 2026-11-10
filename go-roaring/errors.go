package roaring

import "errors"

var (
	// ErrNilBitmap is returned when attempting a set operation against a nil bitmap instance.
	ErrNilBitmap = errors.New("roaring: operand bitmap cannot be nil")

	// ErrCorruptSerialization is returned when deserializing malformed roaring bytes.
	ErrCorruptSerialization = errors.New("roaring: corrupt or malformed serialized bitmap")
)
