package window

import "errors"

var (
	// ErrInvalidDuration is returned when window size or slide interval is <= 0.
	ErrInvalidDuration = errors.New("window: window duration must be greater than zero")

	// ErrLateEventDropped is returned when an event arrives after the watermark allowed lateness.
	ErrLateEventDropped = errors.New("window: event timestamp is too late, dropped by watermark")

	// ErrWindowClosed is returned when operating on a closed window manager.
	ErrWindowClosed = errors.New("window: window manager is closed")
)
