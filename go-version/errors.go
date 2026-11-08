package goversion

import "errors"

var (
	// ErrInvalidVersion indicates a string does not conform to SemVer 2.0.0.
	ErrInvalidVersion = errors.New("goversion: invalid semantic version format")

	// ErrInvalidConstraint indicates an invalid constraint range expression.
	ErrInvalidConstraint = errors.New("goversion: invalid version constraint expression")
)
