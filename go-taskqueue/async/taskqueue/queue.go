package taskqueue

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

type Queue struct {
	cfg      Config
	handlers map[string]HandlerFunc
	taskChan chan *Task
	wg       sync.WaitGroup
	stopChan chan struct{}
	stopOnce sync.Once
	mu       sync.RWMutex
}

func New(cfg Config) (*Queue, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Queue{
		cfg:      cfg,
		handlers: make(map[string]HandlerFunc),
		taskChan: make(chan *Task, cfg.BufferSize),
		stopChan: make(chan struct{}),
	}, nil
}

func (q *Queue) Register(taskType string, fn HandlerFunc) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[taskType] = fn
}

func (q *Queue) Start() {
	for i := 0; i < q.cfg.Workers; i++ {
		q.wg.Add(1)
		go q.workerLoop(i)
	}
}

func (q *Queue) Enqueue(task *Task) error {
	if task == nil {
		return ErrTaskNil
	}

	if task.ID == "" {
		task.ID = generateTaskID()
	}
	now := time.Now()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	task.UpdatedAt = now
	task.Status = StatusPending

	select {
	case q.taskChan <- task:
		return nil
	case <-q.stopChan:
		return ErrQueueClosed
	default:
		return ErrBufferFull
	}
}

func (q *Queue) workerLoop(workerID int) {
	defer q.wg.Done()

	for {
		select {
		case task, ok := <-q.taskChan:
			if !ok {
				return
			}
			q.processTask(task)
		case <-q.stopChan:
			return
		}
	}
}

func (q *Queue) processTask(task *Task) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[taskqueue] panic recovered on task %s (%s): %v", task.ID, task.Type, r)
			task.Status = StatusFailed
			task.LastError = fmt.Sprintf("panic: %v", r)
		}
	}()

	q.mu.RLock()
	fn, exists := q.handlers[task.Type]
	q.mu.RUnlock()

	if !exists {
		task.Status = StatusFailed
		task.LastError = ErrHandlerMissing.Error()
		return
	}

	task.Status = StatusProcessing
	task.UpdatedAt = time.Now()

	err := fn(task)
	if err == nil {
		task.Status = StatusCompleted
		task.UpdatedAt = time.Now()
		return
	}

	task.LastError = err.Error()
	if task.RetryCount < task.MaxRetries {
		task.RetryCount++
		delay := q.cfg.InitialRetryDelay * time.Duration(1<<(task.RetryCount-1))
		if delay > q.cfg.MaxRetryDelay {
			delay = q.cfg.MaxRetryDelay
		}
		time.Sleep(delay)
		_ = q.Enqueue(task)
		return
	}

	task.Status = StatusFailed
	task.UpdatedAt = time.Now()
}

func (q *Queue) Stop() {
	q.stopOnce.Do(func() {
		close(q.stopChan)
		close(q.taskChan)
		q.wg.Wait()
	})
}

func generateTaskID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("tsk_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b[:]))
}
