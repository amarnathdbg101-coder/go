package shadow

import "errors"

var (
	// ErrCandidateUnreachable is reported when the shadowed candidate service fails to connect.
	ErrCandidateUnreachable = errors.New("shadow: candidate target service unreachable")

	// ErrInvalidTargetURL is returned when an invalid URL is supplied for traffic mirroring.
	ErrInvalidTargetURL = errors.New("shadow: invalid target URL")

	// ErrShadowClosed is returned when submitting shadow work to a closed engine.
	ErrShadowClosed = errors.New("shadow: shadow engine is closed")
)
