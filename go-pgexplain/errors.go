package pgexplain

import "errors"

var (
	// ErrInvalidPlanFormat is returned when the PostgreSQL EXPLAIN output cannot be parsed as valid JSON.
	ErrInvalidPlanFormat = errors.New("pgexplain: invalid or unexpected PostgreSQL JSON plan format")

	// ErrNoPlanFound is returned when the EXPLAIN query returns an empty result set.
	ErrNoPlanFound = errors.New("pgexplain: no execution plan found in query output")

	// ErrExplainerClosed is returned when attempting to analyze queries on a closed explainer.
	ErrExplainerClosed = errors.New("pgexplain: explainer has been closed")
)
