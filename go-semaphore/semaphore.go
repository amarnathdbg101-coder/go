package semaphore

import (
	"context"
	"errors"
	"sync"
	"time"
)

type waiter struct {
	n       int64
	ready   chan struct{}
	granted bool
}

// Semaphore implements a weighted, context-aware counting semaphore with FIFO fairness.
type Semaphore struct {
	mu        sync.Mutex
	capacity  int64
	available int64
	waiters   []*waiter
	closed    bool
}

// New creates a new Semaphore with the specified total capacity.
func New(capacity int64) *Semaphore {
	if capacity <= 0 {
		panic("semaphore: capacity must be greater than 0")
	}
	return &Semaphore{
		capacity:  capacity,
		available: capacity,
	}
}

// Acquire acquires n tokens, blocking until available or ctx is cancelled.
func (s *Semaphore) Acquire(ctx context.Context, n int64) error {
	if n <= 0 || n > s.capacity {
		return ErrInvalidWeight
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrClosed
	}

	// Fast path: no waiting queue and sufficient tokens
	if len(s.waiters) == 0 && s.available >= n {
		s.available -= n
		s.mu.Unlock()
		return nil
	}

	// Slow path: enqueue FIFO waiter
	w := &waiter{
		n:     n,
		ready: make(chan struct{}),
	}
	s.waiters = append(s.waiters, w)
	s.mu.Unlock()

	select {
	case <-ctx.Done():
		s.mu.Lock()
		select {
		case <-w.ready:
			if w.granted {
				// Tokens were granted just before cancellation was noticed; refund them
				s.available += n
				s.notifyWaitersLocked()
			}
		default:
			// Remove waiter from queue
			for i, item := range s.waiters {
				if item == w {
					s.waiters = append(s.waiters[:i], s.waiters[i+1:]...)
					break
				}
			}
		}
		s.mu.Unlock()
		return ctx.Err()

	case <-w.ready:
		if !w.granted {
			return ErrClosed
		}
		return nil
	}
}

// AcquireTimeout acquires n tokens, blocking for at most duration d.
func (s *Semaphore) AcquireTimeout(n int64, d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	err := s.Acquire(ctx, n)
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	return err
}

// TryAcquire attempts to acquire n tokens without blocking.
// Returns true if tokens were acquired, false otherwise.
func (s *Semaphore) TryAcquire(n int64) bool {
	if n <= 0 || n > s.capacity {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return false
	}

	if len(s.waiters) == 0 && s.available >= n {
		s.available -= n
		return true
	}

	return false
}

// Release returns n tokens to the semaphore and notifies pending FIFO waiters.
func (s *Semaphore) Release(n int64) error {
	if n <= 0 {
		return ErrInvalidWeight
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.available+n > s.capacity {
		return ErrReleaseExcess
	}

	s.available += n
	s.notifyWaitersLocked()
	return nil
}

// Drain acquires all currently available tokens without blocking and returns the acquired count.
func (s *Semaphore) Drain() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	drained := s.available
	s.available = 0
	return drained
}

// Available returns the number of currently free tokens.
func (s *Semaphore) Available() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.available
}

// Capacity returns the total token capacity of the semaphore.
func (s *Semaphore) Capacity() int64 {
	return s.capacity
}

// Close closes the semaphore, releasing all blocked waiters with ErrClosed.
func (s *Semaphore) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	for _, w := range s.waiters {
		close(w.ready)
	}
	s.waiters = nil
}

func (s *Semaphore) notifyWaitersLocked() {
	for len(s.waiters) > 0 {
		head := s.waiters[0]
		if s.available < head.n {
			break // FIFO fairness: head must be satisfied first
		}
		s.available -= head.n
		s.waiters = s.waiters[1:]
		head.granted = true
		close(head.ready)
	}
}
