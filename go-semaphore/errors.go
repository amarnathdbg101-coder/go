package semaphore

import "errors"

var (
	// ErrTimeout is returned when semaphore acquisition exceeds the specified duration.
	ErrTimeout = errors.New("semaphore: acquisition timed out")

	// ErrInvalidWeight is returned when requesting a weight <= 0 or > total capacity.
	ErrInvalidWeight = errors.New("semaphore: invalid weight (must be > 0 and <= capacity)")

	// ErrReleaseExcess is returned when releasing more tokens than the semaphore's total capacity.
	ErrReleaseExcess = errors.New("semaphore: released more tokens than capacity")

	// ErrClosed is returned when operating on a closed semaphore.
	ErrClosed = errors.New("semaphore: semaphore is closed")
)
