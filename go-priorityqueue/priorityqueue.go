package gopriorityqueue

import (
	"sync"
)

// Order specifies heap ordering (MaxHeap or MinHeap).
type Order int

const (
	MaxHeap Order = iota // Highest priority popped first
	MinHeap              // Lowest priority popped first
)

type entry[T any] struct {
	val      T
	priority int
}

// PriorityQueue implements a thread-safe generic binary heap priority queue.
type PriorityQueue[T any] struct {
	items []entry[T]
	order Order
	mu    sync.RWMutex
}

// New creates a PriorityQueue with default MaxHeap ordering.
func New[T any]() *PriorityQueue[T] {
	return NewWithOrder[T](MaxHeap)
}

// NewWithOrder creates a PriorityQueue with specified ordering (MaxHeap or MinHeap).
func NewWithOrder[T any](order Order) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		items: make([]entry[T], 0),
		order: order,
	}
}

// Push adds an item with an associated priority into the queue in O(log n) time.
func (pq *PriorityQueue[T]) Push(val T, priority int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	item := entry[T]{val: val, priority: priority}
	pq.items = append(pq.items, item)
	pq.siftUp(len(pq.items) - 1)
}

// Pop removes and returns the highest (or lowest) priority item in O(log n) time.
func (pq *PriorityQueue[T]) Pop() (T, int, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	var zero T
	n := len(pq.items)
	if n == 0 {
		return zero, 0, ErrQueueEmpty
	}

	root := pq.items[0]
	last := pq.items[n-1]
	pq.items = pq.items[:n-1]

	if len(pq.items) > 0 {
		pq.items[0] = last
		pq.siftDown(0)
	}

	return root.val, root.priority, nil
}

// Peek returns the top item without removing it in O(1) time.
func (pq *PriorityQueue[T]) Peek() (T, int, error) {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	var zero T
	if len(pq.items) == 0 {
		return zero, 0, ErrQueueEmpty
	}

	top := pq.items[0]
	return top.val, top.priority, nil
}

// Len returns current number of items in the queue.
func (pq *PriorityQueue[T]) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}

// IsEmpty returns true if queue contains no items.
func (pq *PriorityQueue[T]) IsEmpty() bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items) == 0
}

// Clear discards all items in the queue.
func (pq *PriorityQueue[T]) Clear() {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.items = make([]entry[T], 0)
}

// Internal heap operations

func (pq *PriorityQueue[T]) compare(i, j int) bool {
	if pq.order == MaxHeap {
		return pq.items[i].priority > pq.items[j].priority
	}
	return pq.items[i].priority < pq.items[j].priority
}

func (pq *PriorityQueue[T]) siftUp(idx int) {
	for idx > 0 {
		parent := (idx - 1) / 2
		if !pq.compare(idx, parent) {
			break
		}
		pq.items[idx], pq.items[parent] = pq.items[parent], pq.items[idx]
		idx = parent
	}
}

func (pq *PriorityQueue[T]) siftDown(idx int) {
	n := len(pq.items)
	for {
		left := 2*idx + 1
		right := 2*idx + 2
		best := idx

		if left < n && pq.compare(left, best) {
			best = left
		}
		if right < n && pq.compare(right, best) {
			best = right
		}

		if best == idx {
			break
		}

		pq.items[idx], pq.items[best] = pq.items[best], pq.items[idx]
		idx = best
	}
}
