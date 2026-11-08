package taskqueue

import (
	"context"
	"sync"
	"sync/atomic"
)

// Queue defines the storage/buffering contract for task ingestion and dispatch.
type Queue interface {
	// Enqueue pushes a task into the queue. Returns ErrQueueFull if buffer is saturated.
	Enqueue(task *Task) error

	// EnqueueWait pushes a task into the queue, waiting if buffer is full until context is cancelled.
	EnqueueWait(ctx context.Context, task *Task) error

	// Dequeue pulls the next available task, blocking until a task is available or context cancelled.
	Dequeue(ctx context.Context) (*Task, error)

	// Len returns the current number of queued tasks awaiting execution.
	Len() int

	// Close marks the queue as stopped and drains/closes underlying channels.
	Close() error
}

// MemoryQueue is a high-speed, thread-safe, channel-backed in-memory queue.
type MemoryQueue struct {
	ch       chan *Task
	isClosed atomic.Bool
	closeOnce sync.Once
}

// NewMemoryQueue creates a channel-based queue with fixed buffer capacity.
func NewMemoryQueue(bufferSize int) *MemoryQueue {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	return &MemoryQueue{
		ch: make(chan *Task, bufferSize),
	}
}

// Enqueue performs non-blocking push. Returns ErrQueueFull if buffer is saturated.
func (q *MemoryQueue) Enqueue(task *Task) error {
	if q.isClosed.Load() {
		return ErrQueueClosed
	}

	select {
	case q.ch <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

// EnqueueWait pushes task, blocking until buffer space frees up or ctx cancels.
func (q *MemoryQueue) EnqueueWait(ctx context.Context, task *Task) error {
	if q.isClosed.Load() {
		return ErrQueueClosed
	}

	select {
	case q.ch <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Dequeue pulls the next task or waits until available or cancelled.
func (q *MemoryQueue) Dequeue(ctx context.Context) (*Task, error) {
	select {
	case task, ok := <-q.ch:
		if !ok {
			return nil, ErrQueueClosed
		}
		return task, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Len returns current number of tasks waiting in channel buffer.
func (q *MemoryQueue) Len() int {
	return len(q.ch)
}

// Close closes the channel safely once.
func (q *MemoryQueue) Close() error {
	q.closeOnce.Do(func() {
		q.isClosed.Store(true)
		close(q.ch)
	})
	return nil
}
