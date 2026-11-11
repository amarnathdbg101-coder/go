package taskqueue

import "errors"

var (
	ErrTaskNil        = errors.New("taskqueue: task must not be nil")
	ErrQueueClosed    = errors.New("taskqueue: queue is closed")
	ErrHandlerMissing = errors.New("taskqueue: no handler registered for task type")
	ErrBufferFull     = errors.New("taskqueue: task queue buffer is full")
	ErrInvalidConfig  = errors.New("taskqueue: invalid configuration")
)
