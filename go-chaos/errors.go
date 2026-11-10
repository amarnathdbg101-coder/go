package chaos

import "errors"

var (
	// ErrInjectedFault is the default synthetic error returned when an error fault is injected.
	ErrInjectedFault = errors.New("chaos: synthetic fault injected")

	// ErrInjectedTimeout is returned when a synthetic timeout is injected.
	ErrInjectedTimeout = errors.New("chaos: synthetic timeout injected")

	// ErrRuleNotFound is returned when attempting to delete an unmapped rule ID.
	ErrRuleNotFound = errors.New("chaos: rule not found")
)
