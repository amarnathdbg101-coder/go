package gocolor

import "errors"

var (
	// ErrInvalidHEX indicates the hex color string is malformed.
	ErrInvalidHEX = errors.New("gocolor: invalid hex color format (expected #RRGGBB or RRGGBB)")
)
