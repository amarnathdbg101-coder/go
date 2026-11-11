package taskqueue

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// Engine is the central asynchronous task processing orchestrator.
type Engine struct {
	queue     Queue
	handlers  map[string]HandlerFunc
	cfg       Config
	mu        sync.RWMutex
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	isStarted atomic.Bool
	isStopped atomic.Bool
	onFailure func(task *Task, err error)
}

// New creates an initialized Engine backed by the high-speed in-memory queue.
func New(opts ...Option) *Engine {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	queue := NewMemoryQueue(cfg.BufferSize)
	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		queue:    queue,
		handlers: make(map[string]HandlerFunc),
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// NewWithQueue creates an Engine with a custom Queue implementation (e.g. Redis).
func NewWithQueue(q Queue, opts ...Option) *Engine {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		queue:    q,
		handlers: make(map[string]HandlerFunc),
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Register registers a processor function for a given task type identifier (e.g. "email:send").
func (e *Engine) Register(taskType string, handler HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[taskType] = handler
}

// OnFailure sets a global callback hook invoked when a task exhausts all retry attempts.
func (e *Engine) OnFailure(fn func(task *Task, err error)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onFailure = fn
}

// Start launches the background worker pool.
func (e *Engine) Start() {
	if !e.isStarted.CompareAndSwap(false, true) {
		return // Already started
	}

	for i := 0; i < e.cfg.Workers; i++ {
		e.wg.Add(1)
		go e.workerLoop(i)
	}
}

// Enqueue serializes payload and pushes task into queue for immediate background processing.
func (e *Engine) Enqueue(taskType string, payload any) (*Task, error) {
	if e.isStopped.Load() {
		return nil, ErrQueueClosed
	}

	task, err := NewTask(taskType, payload)
	if err != nil {
		return nil, err
	}
	task.MaxRetries = e.cfg.MaxRetries

	if err := e.queue.Enqueue(task); err != nil {
		return nil, err
	}

	return task, nil
}

// EnqueueWait pushes task, waiting if buffer is full until space is available or ctx cancels.
func (e *Engine) EnqueueWait(ctx context.Context, taskType string, payload any) (*Task, error) {
	if e.isStopped.Load() {
		return nil, ErrQueueClosed
	}

	task, err := NewTask(taskType, payload)
	if err != nil {
		return nil, err
	}
	task.MaxRetries = e.cfg.MaxRetries

	if err := e.queue.EnqueueWait(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// EnqueueAfter schedules a task to be enqueued for execution after the specified delay duration.
func (e *Engine) EnqueueAfter(taskType string, payload any, delay time.Duration) (*Task, error) {
	if e.isStopped.Load() {
		return nil, ErrQueueClosed
	}

	task, err := NewTask(taskType, payload)
	if err != nil {
		return nil, err
	}
	task.MaxRetries = e.cfg.MaxRetries

	if delay <= 0 {
		if err := e.queue.Enqueue(task); err != nil {
			return nil, err
		}
		return task, nil
	}

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-e.ctx.Done():
			return
		case <-timer.C:
			if !e.isStopped.Load() {
				_ = e.queue.Enqueue(task)
			}
		}
	}()

	return task, nil
}


// workerLoop is the continuous execution loop for an individual worker goroutine.
func (e *Engine) workerLoop(workerID int) {
	defer e.wg.Done()

	for {
		task, err := e.queue.Dequeue(e.ctx)
		if err != nil {
			return // Context cancelled or queue closed
		}

		e.processTask(task)
	}
}

// processTask handles task dispatching, panic recovery, execution timeout, and exponential retries.
func (e *Engine) processTask(task *Task) {
	// Panic Recovery: Guarantees that a panicking task does not crash the entire worker pool or server process
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[taskqueue] panic recovered on task %s (%s): %v\n", task.ID, task.Type, r)
			e.handleTaskFailure(task, fmt.Errorf("panic: %v", r))
		}
	}()

	e.mu.RLock()
	handler, exists := e.handlers[task.Type]
	e.mu.RUnlock()

	if !exists {
		e.handleTaskFailure(task, ErrHandlerNotFound)
		return
	}

	// Bounded execution timeout per task
	taskCtx, taskCancel := context.WithTimeout(e.ctx, e.cfg.TaskTimeout)
	defer taskCancel()

	err := handler(taskCtx, task)
	if err != nil {
		e.handleTaskFailure(task, err)
		return
	}

	// Task completed successfully
}

// handleTaskFailure manages retry scheduling with exponential backoff.
func (e *Engine) handleTaskFailure(task *Task, taskErr error) {
	task.LastError = taskErr.Error()

	if task.RetryCount < task.MaxRetries {
		task.RetryCount++

		// Calculate exponential backoff delay: initial * 2^(retryCount-1)
		mult := math.Pow(2, float64(task.RetryCount-1))
		backoff := time.Duration(float64(e.cfg.InitialBackoff) * mult)
		if backoff > e.cfg.MaxBackoff {
			backoff = e.cfg.MaxBackoff
		}

		// Re-enqueue after backoff delay in separate goroutine so worker isn't blocked
		go func(t *Task, delay time.Duration) {
			select {
			case <-time.After(delay):
				_ = e.queue.Enqueue(t)
			case <-e.ctx.Done():
				return
			}
		}(task, backoff)

		return
	}

	// Max retries exhausted
	e.mu.RLock()
	hook := e.onFailure
	e.mu.RUnlock()

	if hook != nil {
		hook(task, taskErr)
	}
}

// Shutdown initiates a graceful stop:
// 1. Stops accepting new tasks and closes queue.
// 2. Waits for all active workers to finish currently running tasks, bounded by shutdownCtx.
func (e *Engine) Shutdown(shutdownCtx context.Context) error {
	if !e.isStopped.CompareAndSwap(false, true) {
		return nil // Already stopped
	}

	// Close queue
	_ = e.queue.Close()

	// Wait for workers in goroutine
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		e.cancel()
		return nil
	case <-shutdownCtx.Done():
		e.cancel() // Force stop on timeout
		return shutdownCtx.Err()
	}
}

// Queue returns the underlying Queue interface.
func (e *Engine) Queue() Queue {
	return e.queue
}
