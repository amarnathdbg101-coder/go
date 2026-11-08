package gofsm

import (
	"fmt"
	"sync"
)

type transitionKey[S, E comparable] struct {
	from  S
	event E
}

type transition[S, E comparable] struct {
	to    S
	guard func() bool
}

// EnterHook is invoked when entering a specific state.
type EnterHook[S, E comparable] func(from S, event E)

// ExitHook is invoked when exiting a specific state.
type ExitHook[S, E comparable] func(to S, event E)

// TransitionHook is invoked on every valid state transition.
type TransitionHook[S, E comparable] func(from S, event E, to S)

// FSM is a thread-safe generic Finite State Machine.
type FSM[S comparable, E comparable] struct {
	current         S
	transitions     map[transitionKey[S, E]]transition[S, E]
	enterHooks      map[S][]EnterHook[S, E]
	exitHooks       map[S][]ExitHook[S, E]
	transitionHooks []TransitionHook[S, E]
	mu              sync.RWMutex
}

// New creates an FSM initialized to the given starting state.
func New[S comparable, E comparable](initial S) *FSM[S, E] {
	return &FSM[S, E]{
		current:     initial,
		transitions: make(map[transitionKey[S, E]]transition[S, E]),
		enterHooks:  make(map[S][]EnterHook[S, E]),
		exitHooks:   make(map[S][]ExitHook[S, E]),
	}
}

// AddTransition registers a transition from one state to another triggered by an event.
// An optional guard predicate can be supplied (pass nil if no guard).
func (f *FSM[S, E]) AddTransition(from S, event E, to S, guard func() bool) *FSM[S, E] {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := transitionKey[S, E]{from: from, event: event}
	f.transitions[key] = transition[S, E]{to: to, guard: guard}
	return f
}

// OnEnter registers a callback executed when entering state S.
func (f *FSM[S, E]) OnEnter(state S, hook EnterHook[S, E]) *FSM[S, E] {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.enterHooks[state] = append(f.enterHooks[state], hook)
	return f
}

// OnExit registers a callback executed when exiting state S.
func (f *FSM[S, E]) OnExit(state S, hook ExitHook[S, E]) *FSM[S, E] {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.exitHooks[state] = append(f.exitHooks[state], hook)
	return f
}

// OnTransition registers a global callback executed on every transition.
func (f *FSM[S, E]) OnTransition(hook TransitionHook[S, E]) *FSM[S, E] {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.transitionHooks = append(f.transitionHooks, hook)
	return f
}

// Current returns the current state.
func (f *FSM[S, E]) Current() S {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current
}

// Can checks whether the event is permissible from the current state.
func (f *FSM[S, E]) Can(event E) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	key := transitionKey[S, E]{from: f.current, event: event}
	tr, exists := f.transitions[key]
	if !exists {
		return false
	}
	if tr.guard != nil && !tr.guard() {
		return false
	}
	return true
}

// Trigger fires an event, validating guards and updating state atomically.
func (f *FSM[S, E]) Trigger(event E) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := transitionKey[S, E]{from: f.current, event: event}
	tr, exists := f.transitions[key]
	if !exists {
		return fmt.Errorf("%w: state %v + event %v", ErrInvalidTransition, f.current, event)
	}

	if tr.guard != nil && !tr.guard() {
		return ErrGuardFailed
	}

	oldState := f.current
	newState := tr.to

	// 1. Fire Exit Hooks
	for _, hook := range f.exitHooks[oldState] {
		hook(newState, event)
	}

	// 2. Mutate State
	f.current = newState

	// 3. Fire Enter Hooks
	for _, hook := range f.enterHooks[newState] {
		hook(oldState, event)
	}

	// 4. Fire Global Transition Hooks
	for _, hook := range f.transitionHooks {
		hook(oldState, event, newState)
	}

	return nil
}

// AvailableEvents returns all events valid from the current state.
func (f *FSM[S, E]) AvailableEvents() []E {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var events []E
	for key, tr := range f.transitions {
		if key.from == f.current {
			if tr.guard == nil || tr.guard() {
				events = append(events, key.event)
			}
		}
	}
	return events
}
