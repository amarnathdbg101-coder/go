package gopriorityqueue

import "errors"

var (
	// ErrQueueEmpty indicates pop or peek was attempted on an empty queue.
	ErrQueueEmpty = errors.New("gopriorityqueue: priority queue is empty")
)
