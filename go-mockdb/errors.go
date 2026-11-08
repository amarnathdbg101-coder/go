package mockdb

import "errors"

var (
	// ErrUnexpectedQuery is returned when a query executes without a matching expectation.
	ErrUnexpectedQuery = errors.New("mockdb: unexpected query executed")
	// ErrUnmetExpectations is returned when registered expectations were not satisfied.
	ErrUnmetExpectations = errors.New("mockdb: there are remaining unmet expectations")
	// ErrArgumentMismatch is returned when query arguments do not match expected arguments.
	ErrArgumentMismatch = errors.New("mockdb: query argument mismatch")
)
