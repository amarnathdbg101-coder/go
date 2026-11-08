package goringbuffer

import (
	"sync"
)

// RingBuffer implements a thread-safe circular FIFO buffer of fixed capacity.
type RingBuffer[T any] struct {
	buf       []T
	head      int
	tail      int
	count     int
	capacity  int
	overwrite bool
	mu        sync.RWMutex
}

// Config specifies ring buffer options.
type Config struct {
	Capacity  int
	Overwrite bool
}

// New creates a RingBuffer with specified capacity and overwrite behavior.
func New[T any](capacity int, overwrite bool) (*RingBuffer[T], error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &RingBuffer[T]{
		buf:       make([]T, capacity),
		capacity:  capacity,
		overwrite: overwrite,
	}, nil
}

// Push appends an item to the buffer.
// If full and overwrite is false, returns ErrBufferFull.
// If full and overwrite is true, overwrites the oldest element.
func (r *RingBuffer[T]) Push(item T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.count == r.capacity {
		if !r.overwrite {
			return ErrBufferFull
		}
		// Overwrite oldest item: advance tail
		r.buf[r.head] = item
		r.head = (r.head + 1) % r.capacity
		r.tail = (r.tail + 1) % r.capacity
		return nil
	}

	r.buf[r.head] = item
	r.head = (r.head + 1) % r.capacity
	r.count++
	return nil
}

// Pop removes and returns the oldest item in FIFO order.
func (r *RingBuffer[T]) Pop() (T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var zero T
	if r.count == 0 {
		return zero, ErrBufferEmpty
	}

	item := r.buf[r.tail]
	r.buf[r.tail] = zero // Clear reference for GC
	r.tail = (r.tail + 1) % r.capacity
	r.count--

	return item, nil
}

// Peek returns the oldest item without removing it.
func (r *RingBuffer[T]) Peek() (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var zero T
	if r.count == 0 {
		return zero, ErrBufferEmpty
	}

	return r.buf[r.tail], nil
}

// Slice returns all active items in FIFO order without mutating buffer.
func (r *RingBuffer[T]) Slice() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]T, r.count)
	for i := 0; i < r.count; i++ {
		idx := (r.tail + i) % r.capacity
		result[i] = r.buf[idx]
	}
	return result
}

// Len returns the current number of buffered elements.
func (r *RingBuffer[T]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.count
}

// Cap returns the total capacity of the buffer.
func (r *RingBuffer[T]) Cap() int {
	return r.capacity
}

// IsFull returns true if count == capacity.
func (r *RingBuffer[T]) IsFull() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.count == r.capacity
}

// IsEmpty returns true if count == 0.
func (r *RingBuffer[T]) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.count == 0
}

// Reset clears the buffer and discards all items.
func (r *RingBuffer[T]) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	var zero T
	for i := range r.buf {
		r.buf[i] = zero
	}
	r.head = 0
	r.tail = 0
	r.count = 0
}
