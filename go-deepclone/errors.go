package deepclone

import "errors"

var (
	// ErrUnsupportedKind is returned when attempting to clone an unsupported type like channel or unsafe pointer.
	ErrUnsupportedKind = errors.New("deepclone: unsupported kind for cloning")
)
