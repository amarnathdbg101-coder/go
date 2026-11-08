package gostopwatch

import "errors"

var (
	// ErrAlreadyStopped indicates Stop() was called on an already stopped stopwatch.
	ErrAlreadyStopped = errors.New("gostopwatch: stopwatch already stopped")
)
