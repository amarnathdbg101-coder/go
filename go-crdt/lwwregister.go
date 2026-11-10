package crdt

import (
	"sync"
	"time"
)

// LWWRegister implements a Last-Write-Wins Register CRDT using Lamport or wall-clock timestamps
// with deterministic replica-ID tie-breaking.
type LWWRegister[T any] struct {
	mu        sync.RWMutex
	nodeID    string
	value     T
	timestamp int64
}

// NewLWWRegister creates a new Last-Write-Wins Register initialized with initialValue.
func NewLWWRegister[T any](nodeID string, initialValue T) *LWWRegister[T] {
	if nodeID == "" {
		nodeID = "node_default"
	}
	return &LWWRegister[T]{
		nodeID:    nodeID,
		value:     initialValue,
		timestamp: time.Now().UnixNano(),
	}
}

// Set updates the register value and stamps it with the current Unix nanosecond timestamp.
func (r *LWWRegister[T]) Set(val T) {
	r.SetWithClock(val, time.Now().UnixNano())
}

// SetWithClock updates the register value using an explicit logical or hybrid timestamp.
func (r *LWWRegister[T]) SetWithClock(val T, ts int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Enforce monotonic timestamp progression on local node
	if ts <= r.timestamp {
		ts = r.timestamp + 1
	}

	r.value = val
	r.timestamp = ts
}

// Get returns the current value and its timestamp.
func (r *LWWRegister[T]) Get() (T, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value, r.timestamp
}

// Value returns only the active value.
func (r *LWWRegister[T]) Value() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// Timestamp returns the active timestamp.
func (r *LWWRegister[T]) Timestamp() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.timestamp
}

// NodeID returns the replica node ID.
func (r *LWWRegister[T]) NodeID() string {
	return r.nodeID
}

// Merge converges this register with another register.
// If other.timestamp > this.timestamp, other wins.
// If timestamps tie, the node with lexicographically greater NodeID wins deterministically.
func (r *LWWRegister[T]) Merge(other *LWWRegister[T]) error {
	if other == nil {
		return ErrNilNode
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	if other.timestamp > r.timestamp {
		r.value = other.value
		r.timestamp = other.timestamp
		r.nodeID = other.nodeID
	} else if other.timestamp == r.timestamp {
		// Deterministic tie-breaker
		if other.nodeID > r.nodeID {
			r.value = other.value
			r.nodeID = other.nodeID
		}
	}

	return nil
}
