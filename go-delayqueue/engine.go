package delayqueue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// HandlerFunc is the signature for handling a delayed job payload.
type HandlerFunc func(ctx context.Context, job *Job) error

// Queue manages scheduling, polling, and dispatching of delayed jobs.
type Queue struct {
	driver   Driver
	cfg      Config
	handlers map[string]HandlerFunc

	mu       sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
	closed   bool
}

// New creates a new delayed job queue instance.
func New(driver Driver, opts ...Option) *Queue {
	if driver == nil {
		driver = NewMemoryDriver()
	}

	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Queue{
		driver:   driver,
		cfg:      cfg,
		handlers: make(map[string]HandlerFunc),
		stopChan: make(chan struct{}),
	}
}

// Register registers a handler function for a specific topic.
func (q *Queue) Register(topic string, handler HandlerFunc) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[topic] = handler
}

// Enqueue schedules a job to execute after the specified delay.
func (q *Queue) Enqueue(topic string, payload []byte, delay time.Duration) (string, error) {
	if delay < 0 {
		return "", ErrInvalidDelay
	}

	id := generateID()
	now := time.Now()
	job := &Job{
		ID:             id,
		Topic:          topic,
		Payload:        payload,
		ExecuteAt:      now.Add(delay),
		MaxRetries:     q.cfg.DefaultMaxRetries,
		CurrentRetries: 0,
		Status:         StatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := q.driver.Push(job); err != nil {
		return "", err
	}
	return id, nil
}

// Cancel cancels a scheduled job.
func (q *Queue) Cancel(id string) error {
	return q.driver.Cancel(id)
}

// Start begins the background poller and worker loop.
func (q *Queue) Start() error {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return nil
	}
	if q.closed {
		q.mu.Unlock()
		return ErrQueueClosed
	}
	q.running = true
	q.mu.Unlock()

	jobChan := make(chan *Job, q.cfg.BatchSize)

	// Launch worker pool
	for i := 0; i < q.cfg.WorkerCount; i++ {
		q.wg.Add(1)
		go q.worker(jobChan)
	}

	// Launch poller loop
	q.wg.Add(1)
	go q.poller(jobChan)

	return nil
}

// Stop gracefully shuts down the queue poller and workers.
func (q *Queue) Stop() error {
	q.mu.Lock()
	if !q.running || q.closed {
		q.mu.Unlock()
		return nil
	}
	q.closed = true
	q.running = false
	close(q.stopChan)
	q.mu.Unlock()

	q.wg.Wait()
	return q.driver.Close()
}

func (q *Queue) poller(jobChan chan<- *Job) {
	defer q.wg.Done()
	defer close(jobChan)

	ticker := time.NewTicker(q.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopChan:
			return
		case now := <-ticker.C:
			dueJobs, err := q.driver.PopDue(now, q.cfg.BatchSize)
			if err != nil {
				continue
			}

			for _, job := range dueJobs {
				select {
				case <-q.stopChan:
					return
				case jobChan <- job:
				}
			}
		}
	}
}

func (q *Queue) worker(jobChan <-chan *Job) {
	defer q.wg.Done()

	for job := range jobChan {
		q.processJob(job)
	}
}

func (q *Queue) processJob(job *Job) {
	q.mu.RLock()
	handler, exists := q.handlers[job.Topic]
	q.mu.RUnlock()

	if !exists {
		_ = q.driver.Nack(job.ID, time.Now().Add(q.cfg.RetryBackoff), ErrNoHandler)
		return
	}

	ctx := context.Background()
	err := handler(ctx, job)
	if err == nil {
		_ = q.driver.Ack(job.ID)
		return
	}

	// Exponential backoff calculation
	retryFactor := time.Duration(1 << job.CurrentRetries)
	nextRetry := time.Now().Add(q.cfg.RetryBackoff * retryFactor)
	_ = q.driver.Nack(job.ID, nextRetry, err)
}

func generateID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("job_%s", hex.EncodeToString(b[:]))
}
