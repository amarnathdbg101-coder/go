package gohttpclient

import (
	"sync"
	"time"
)

// CircuitState represents the current health state of the circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreaker prevents cascading failures when downstream services fail.
type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	consecutiveFails int
	maxFailures      int
	cooldown         time.Duration
	openedAt         time.Time
}

// NewCircuitBreaker creates a circuit breaker with failure threshold and cooldown.
func NewCircuitBreaker(maxFailures int, cooldown time.Duration) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if cooldown <= 0 {
		cooldown = 10 * time.Second
	}
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		cooldown:    cooldown,
	}
}

// AllowRequest checks whether a request can be dispatched.
func (cb *CircuitBreaker) AllowRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		if now.Sub(cb.openedAt) >= cb.cooldown {
			// Cooldown expired, enter HalfOpen state for canary probe
			cb.state = StateHalfOpen
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// In HalfOpen, allow single probe (already permitted once)
		return nil
	}

	return nil
}

// RecordSuccess records a successful execution and resets the breaker to Closed.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFails = 0
	cb.state = StateClosed
}

// RecordFailure records a failed request and trips the breaker if threshold reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFails++

	if cb.state == StateHalfOpen || cb.consecutiveFails >= cb.maxFailures {
		cb.state = StateOpen
		cb.openedAt = time.Now()
	}
}

// State returns current breaker state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
