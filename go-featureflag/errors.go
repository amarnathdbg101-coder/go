package featureflag

import "errors"

var (
	// ErrFlagNotFound is returned when querying a flag that is not registered.
	ErrFlagNotFound = errors.New("featureflag: flag not found")

	// ErrInvalidPercentage is returned when rollout percentage is outside 0-100 range.
	ErrInvalidPercentage = errors.New("featureflag: rollout percentage must be between 0 and 100")
)
