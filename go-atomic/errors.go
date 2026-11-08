package atomic

import "errors"

var (
	// ErrNilValue is returned when storing a nil value into a typed atomic container where disallowed.
	ErrNilValue = errors.New("atomic: cannot store nil value")
)
