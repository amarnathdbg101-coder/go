package gosqlbuilder

import "errors"

var (
	// ErrNoTable indicates a query was built without specifying a table name.
	ErrNoTable = errors.New("sqlbuilder: table name is required")

	// ErrNoColumns indicates an INSERT or SELECT has no target columns.
	ErrNoColumns = errors.New("sqlbuilder: at least one column is required")

	// ErrNoValues indicates an INSERT statement has no values.
	ErrNoValues = errors.New("sqlbuilder: no values provided for insert")

	// ErrMismatchedValues indicates the count of values does not match column count in INSERT.
	ErrMismatchedValues = errors.New("sqlbuilder: column count and value count mismatch")

	// ErrNoSetClauses indicates an UPDATE statement has no SET columns.
	ErrNoSetClauses = errors.New("sqlbuilder: update statement requires at least one SET clause")

	// ErrInvalidOrderBy indicates an invalid order direction (must be ASC or DESC).
	ErrInvalidOrderBy = errors.New("sqlbuilder: order direction must be ASC or DESC")
)
