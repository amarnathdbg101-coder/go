package taskqueue_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-taskqueue/async/taskqueue"
)

func TestTaskQueue_BasicProcess(t *testing.T) {
	q, err := taskqueue.New(taskqueue.Config{Workers: 2, BufferSize: 10})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	q.Start()
	defer q.Stop()

	var processed int64
	q.Register("send_email", func(task *taskqueue.Task) error {
		atomic.AddInt64(&processed, 1)
		return nil
	})

	err = q.Enqueue(&taskqueue.Task{
		Type:    "send_email",
		Payload: map[string]interface{}{"to": "user@example.com"},
	})
	if err != nil {
		t.Fatalf("failed to enqueue task: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt64(&processed) != 1 {
		t.Errorf("expected 1 task processed, got %d", atomic.LoadInt64(&processed))
	}
}

func TestTaskQueue_ConcurrentWorkers(t *testing.T) {
	q, _ := taskqueue.New(taskqueue.Config{Workers: 10, BufferSize: 100})
	q.Start()
	defer q.Stop()

	var counter int64
	const totalTasks = 50

	q.Register("increment", func(task *taskqueue.Task) error {
		atomic.AddInt64(&counter, 1)
		return nil
	})

	var wg sync.WaitGroup
	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = q.Enqueue(&taskqueue.Task{
				Type:    "increment",
				Payload: map[string]interface{}{"id": id},
			})
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&counter) != totalTasks {
		t.Errorf("expected %d tasks processed, got %d", totalTasks, atomic.LoadInt64(&counter))
	}
}

func TestTaskQueue_PanicRecovery(t *testing.T) {
	q, _ := taskqueue.New(taskqueue.Config{Workers: 1, BufferSize: 10})
	q.Start()
	defer q.Stop()

	q.Register("panicking", func(task *taskqueue.Task) error {
		panic("critical unexpected database disconnection!")
	})

	err := q.Enqueue(&taskqueue.Task{
		Type: "panicking",
	})
	if err != nil {
		t.Fatalf("failed to enqueue panicking task: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
}
