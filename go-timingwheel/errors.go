package timingwheel

import "errors"

var (
	// ErrInvalidTick is returned when tick duration or wheel size is <= 0.
	ErrInvalidTick = errors.New("timingwheel: tick duration and wheel size must be greater than 0")

	// ErrWheelStopped is returned when attempting to schedule a task on a stopped timing wheel.
	ErrWheelStopped = errors.New("timingwheel: timing wheel is stopped")
)
