package queryfilter

import "errors"

var (
	// ErrDisallowedField is returned when a query specifies a filter/sort field not in the whitelist.
	ErrDisallowedField = errors.New("queryfilter: field not permitted by whitelist")
	// ErrInvalidOperator is returned when an unsupported comparison operator is requested.
	ErrInvalidOperator = errors.New("queryfilter: unsupported filter operator")
	// ErrNilURLValues is returned when url.Values is nil.
	ErrNilURLValues = errors.New("queryfilter: query values cannot be nil")
)
