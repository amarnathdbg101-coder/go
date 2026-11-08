package gofsm

import "errors"

var (
	// ErrInvalidTransition indicates attempting an event with no matching transition rule from current state.
	ErrInvalidTransition = errors.New("gofsm: invalid state transition for given event")

	// ErrGuardFailed indicates transition rule matched but guard predicate returned false.
	ErrGuardFailed = errors.New("gofsm: transition guard condition failed")
)
