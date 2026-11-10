package delayqueue_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-delayqueue"
)

type OrderCancelledEvent struct {
	OrderID string `json:"order_id"`
}

func TestDelayedExecution(t *testing.T) {
	driver := delayqueue.NewMemoryDriver()
	q := delayqueue.New(driver, delayqueue.WithPollInterval(10*time.Millisecond))

	var executed int32
	q.Register("order.auto_cancel", func(ctx context.Context, job *delayqueue.Job) error {
		atomic.AddInt32(&executed, 1)
		return nil
	})

	if err := q.Start(); err != nil {
		t.Fatalf("failed to start queue: %v", err)
	}
	defer q.Stop()

	// Enqueue job with 30ms delay
	_, err := q.Enqueue("order.auto_cancel", []byte("order_999"), 30*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	// At 10ms, it should NOT have executed yet
	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&executed) != 0 {
		t.Fatalf("job executed prematurely before delay!")
	}

	// Wait past the 30ms threshold
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&executed) != 1 {
		t.Fatalf("expected job to execute after delay, got count %d", atomic.LoadInt32(&executed))
	}
}

func TestCancelJob(t *testing.T) {
	driver := delayqueue.NewMemoryDriver()
	q := delayqueue.New(driver, delayqueue.WithPollInterval(10*time.Millisecond))

	var executed int32
	q.Register("cart.abandoned", func(ctx context.Context, job *delayqueue.Job) error {
		atomic.AddInt32(&executed, 1)
		return nil
	})

	_ = q.Start()
	defer q.Stop()

	jobID, err := q.Enqueue("cart.abandoned", []byte("cart_123"), 40*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	// Cancel before due time
	if err := q.Cancel(jobID); err != nil {
		t.Fatalf("failed to cancel: %v", err)
	}

	time.Sleep(60 * time.Millisecond)
	if atomic.LoadInt32(&executed) != 0 {
		t.Fatalf("cancelled job should never execute!")
	}
}

func TestRetryAndDeadLetter(t *testing.T) {
	driver := delayqueue.NewMemoryDriver()
	q := delayqueue.New(
		driver,
		delayqueue.WithPollInterval(5*time.Millisecond),
		delayqueue.WithDefaultMaxRetries(2),
		delayqueue.WithRetryBackoff(5*time.Millisecond),
	)

	var attempts int32
	testErr := errors.New("simulated network glitch")

	q.Register("webhook.dispatch", func(ctx context.Context, job *delayqueue.Job) error {
		atomic.AddInt32(&attempts, 1)
		return testErr
	})

	_ = q.Start()
	defer q.Stop()

	jobID, err := q.Enqueue("webhook.dispatch", []byte("hook_abc"), 2*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	// Allow initial execution and retries to occur
	time.Sleep(100 * time.Millisecond)

	job, err := driver.Get(jobID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}

	if job.Status != delayqueue.StatusDeadLetter {
		t.Fatalf("expected job to end in DEAD_LETTER, got %s", job.Status)
	}
}

func TestEnqueueJSON(t *testing.T) {
	driver := delayqueue.NewMemoryDriver()
	q := delayqueue.New(driver)

	event := OrderCancelledEvent{OrderID: "ORD-2026-X"}
	id, err := delayqueue.EnqueueJSON(q, "order.cancelled", event, 1*time.Minute)
	if err != nil || id == "" {
		t.Fatalf("failed to enqueue JSON: %v", err)
	}

	job, err := driver.Get(id)
	if err != nil || string(job.Payload) != `{"order_id":"ORD-2026-X"}` {
		t.Fatalf("unexpected payload: %s", string(job.Payload))
	}
}

