package latch

import (
	"context"
	"sync"
	"time"
)

// CountDownLatch allows one or more goroutines to wait until a set of operations completed.
type CountDownLatch struct {
	mu    sync.Mutex
	count int
	done  chan struct{}
}

// NewCountDownLatch creates a CountDownLatch initialized with the given count.
func NewCountDownLatch(count int) *CountDownLatch {
	if count < 0 {
		count = 0
	}
	cdl := &CountDownLatch{
		count: count,
		done:  make(chan struct{}),
	}
	if count == 0 {
		close(cdl.done)
	}
	return cdl
}

// CountDown decrements the count of the latch, releasing all waiting goroutines when count reaches zero.
func (c *CountDownLatch) CountDown() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.count > 0 {
		c.count--
		if c.count == 0 {
			close(c.done)
		}
	}
}

// GetCount returns the current count.
func (c *CountDownLatch) GetCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// Wait blocks until the latch count reaches zero.
func (c *CountDownLatch) Wait() {
	<-c.done
}

// WaitTimeout blocks until the latch count reaches zero or duration d elapses.
// Returns true if count reached zero, false if timed out.
func (c *CountDownLatch) WaitTimeout(d time.Duration) bool {
	select {
	case <-c.done:
		return true
	case <-time.After(d):
		return false
	}
}

// WaitContext blocks until the latch count reaches zero or ctx is cancelled.
func (c *CountDownLatch) WaitContext(ctx context.Context) error {
	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// CyclicBarrier coordinates a fixed number of goroutines that must all wait for each other at a common barrier.
type CyclicBarrier struct {
	mu      sync.Mutex
	parties int
	count   int
	action  func()
	gen     *generation
}

type generation struct {
	broken bool
	done   chan struct{}
}

// NewCyclicBarrier creates a new CyclicBarrier for the given number of parties, with an optional barrier action.
func NewCyclicBarrier(parties int, action func()) *CyclicBarrier {
	if parties <= 0 {
		panic("latch: parties must be greater than 0")
	}
	return &CyclicBarrier{
		parties: parties,
		count:   parties,
		action:  action,
		gen:     &generation{done: make(chan struct{})},
	}
}

// Await waits until all parties have invoked Await on this barrier.
// Returns the arrival index (parties-1 down to 0).
func (b *CyclicBarrier) Await(ctx context.Context) (int, error) {
	b.mu.Lock()
	g := b.gen

	if g.broken {
		b.mu.Unlock()
		return 0, ErrBrokenBarrier
	}

	b.count--
	index := b.count

	// If this is the last party to arrive:
	if b.count == 0 {
		if b.action != nil {
			b.action()
		}
		// Complete current generation and reset for the next cycle
		close(g.done)
		b.count = b.parties
		b.gen = &generation{done: make(chan struct{})}
		b.mu.Unlock()
		return index, nil
	}

	b.mu.Unlock()

	select {
	case <-ctx.Done():
		b.mu.Lock()
		if b.gen == g && !g.broken {
			g.broken = true
			close(g.done)
		}
		b.mu.Unlock()
		return 0, ctx.Err()

	case <-g.done:
		b.mu.Lock()
		defer b.mu.Unlock()
		if g.broken {
			return 0, ErrBrokenBarrier
		}
		return index, nil
	}
}

// Reset resets the barrier to its initial state, breaking any currently waiting generation.
func (b *CyclicBarrier) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.gen.broken {
		b.gen.broken = true
		close(b.gen.done)
	}
	b.count = b.parties
	b.gen = &generation{done: make(chan struct{})}
}

// Parties returns the required number of parties to trip the barrier.
func (b *CyclicBarrier) Parties() int {
	return b.parties
}

// NumberWaiting returns the number of parties currently waiting at the barrier.
func (b *CyclicBarrier) NumberWaiting() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.parties - b.count
}

// IsBroken returns true if this barrier is in a broken state.
func (b *CyclicBarrier) IsBroken() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.gen.broken
}
