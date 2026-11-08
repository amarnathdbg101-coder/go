package slidingstats

import "errors"

var (
	// ErrInvalidWindow is returned when window duration is non-positive.
	ErrInvalidWindow = errors.New("slidingstats: window duration must be positive")
	// ErrInvalidBuckets is returned when bucket count is less than 1.
	ErrInvalidBuckets = errors.New("slidingstats: bucket count must be at least 1")
	// ErrNoData is returned when computing statistics on an empty window.
	ErrNoData = errors.New("slidingstats: no data points in current window")
)
