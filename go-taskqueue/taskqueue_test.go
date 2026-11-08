package taskqueue_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-taskqueue"
)

type WelcomeEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// -------------------------------------------------------------------------
// 1. Basic Processing & Payload Unmarshaling Tests
// -------------------------------------------------------------------------

func TestTaskQueue_BasicProcess(t *testing.T) {
	engine := taskqueue.New(
		taskqueue.WithWorkers(2),
		taskqueue.WithBufferSize(50),
	)

	var processedEmail string
	var wg sync.WaitGroup
	wg.Add(1)

	engine.Register("email:welcome", func(ctx context.Context, task *taskqueue.Task) error {
		var payload WelcomeEmailPayload
		if err := task.UnmarshalPayload(&payload); err != nil {
			return err
		}
		processedEmail = payload.Email
		wg.Done()
		return nil
	})

	engine.Start()
	defer func() {
		_ = engine.Shutdown(context.Background())
	}()

	task, err := engine.Enqueue("email:welcome", WelcomeEmailPayload{
		UserID: "usr_100",
		Email:  "user@example.com",
	})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	if task.ID == "" {
		t.Errorf("expected non-empty task ID")
	}

	wg.Wait()

	if processedEmail != "user@example.com" {
		t.Errorf("expected 'user@example.com', got '%s'", processedEmail)
	}
}

// -------------------------------------------------------------------------
// 2. High Concurrency Worker Pool Tests
// -------------------------------------------------------------------------

func TestTaskQueue_ConcurrentWorkers(t *testing.T) {
	engine := taskqueue.New(
		taskqueue.WithWorkers(4),
		taskqueue.WithBufferSize(200),
	)

	const totalTasks = 50
	var completedCount int64
	var wg sync.WaitGroup
	wg.Add(totalTasks)

	engine.Register("math:increment", func(ctx context.Context, task *taskqueue.Task) error {
		// Simulate small work
		time.Sleep(5 * time.Millisecond)
		atomic.AddInt64(&completedCount, 1)
		wg.Done()
		return nil
	})

	engine.Start()
	defer func() {
		_ = engine.Shutdown(context.Background())
	}()

	for i := 0; i < totalTasks; i++ {
		_, err := engine.Enqueue("math:increment", map[string]int{"val": i})
		if err != nil {
			t.Fatalf("enqueue error on task %d: %v", i, err)
		}
	}

	wg.Wait()

	if atomic.LoadInt64(&completedCount) != totalTasks {
		t.Errorf("expected %d completed tasks, got %d", totalTasks, completedCount)
	}
}

// -------------------------------------------------------------------------
// 3. Panic Recovery Tests (Zero Process Crash)
// -------------------------------------------------------------------------

func TestTaskQueue_PanicRecovery(t *testing.T) {
	engine := taskqueue.New(
		taskqueue.WithWorkers(2),
		taskqueue.WithMaxRetries(0), // No retries for immediate failure check
	)

	var failedTaskID string
	var wg sync.WaitGroup
	wg.Add(1)

	engine.OnFailure(func(task *taskqueue.Task, err error) {
		failedTaskID = task.ID
		wg.Done()
	})

	// Panicking handler
	engine.Register("task:panicking", func(ctx context.Context, task *taskqueue.Task) error {
		panic("critical unexpected database disconnection!")
	})

	engine.Start()
	defer func() {
		_ = engine.Shutdown(context.Background())
	}()

	task, err := engine.Enqueue("task:panicking", "panic payload")
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	wg.Wait()

	if failedTaskID != task.ID {
		t.Errorf("expected failed task ID '%s', got '%s'", task.ID, failedTaskID)
	}
}

// -------------------------------------------------------------------------
// 4. Retry with Exponential Backoff Tests
// -------------------------------------------------------------------------

func TestTaskQueue_RetrySuccess(t *testing.T) {
	engine := taskqueue.New(
		taskqueue.WithWorkers(2),
		taskqueue.WithMaxRetries(3),
		taskqueue.WithBackoff(10*time.Millisecond, 50*time.Millisecond),
	)

	var attempts int64
	var wg sync.WaitGroup
	wg.Add(1)

	engine.Register("task:flaky", func(ctx context.Context, task *taskqueue.Task) error {
		count := atomic.AddInt64(&attempts, 1)
		if count < 3 {
			// Fail first 2 attempts
			return errors.New("temporary network glitch")
		}
		// Succeed on 3rd attempt
		wg.Done()
		return nil
	})

	engine.Start()
	defer func() {
		_ = engine.Shutdown(context.Background())
	}()

	_, err := engine.Enqueue("task:flaky", "test")
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	wg.Wait()

	if atomic.LoadInt64(&attempts) != 3 {
		t.Errorf("expected exactly 3 attempts, got %d", attempts)
	}
}

// -------------------------------------------------------------------------
// 5. Graceful Shutdown & Buffer Tests
// -------------------------------------------------------------------------

func TestTaskQueue_GracefulShutdown(t *testing.T) {
	engine := taskqueue.New(
		taskqueue.WithWorkers(2),
		taskqueue.WithBufferSize(10),
	)

	var processed int64
	engine.Register("task:slow", func(ctx context.Context, task *taskqueue.Task) error {
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt64(&processed, 1)
		return nil
	})

	engine.Start()

	// Enqueue tasks
	_, _ = engine.Enqueue("task:slow", "1")
	_, _ = engine.Enqueue("task:slow", "2")

	// Shutdown should wait for tasks 1 and 2 to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := engine.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}

	if atomic.LoadInt64(&processed) != 2 {
		t.Errorf("expected 2 tasks to complete during graceful shutdown, got %d", processed)
	}

	// Enqueue after shutdown should return ErrQueueClosed
	_, errAfter := engine.Enqueue("task:slow", "3")
	if errAfter != taskqueue.ErrQueueClosed {
		t.Errorf("expected ErrQueueClosed after shutdown, got: %v", errAfter)
	}
}

func TestTaskQueue_BufferFull(t *testing.T) {
	q := taskqueue.NewMemoryQueue(1) // Capacity: 1

	task1, _ := taskqueue.NewTask("t1", nil)
	task2, _ := taskqueue.NewTask("t2", nil)

	if err := q.Enqueue(task1); err != nil {
		t.Fatalf("first enqueue failed: %v", err)
	}

	// Buffer full
	if err := q.Enqueue(task2); err != taskqueue.ErrQueueFull {
		t.Errorf("expected ErrQueueFull on saturated buffer, got: %v", err)
	}
}
