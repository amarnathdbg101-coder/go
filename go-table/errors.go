package table

import "errors"

var (
	// ErrEmptyTable is returned when attempting an operation requiring table rows on an empty table.
	ErrEmptyTable = errors.New("table: table has no rows")
)
