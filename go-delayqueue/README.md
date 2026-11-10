# go-delayqueue ⏱️

> **Production-Grade Delayed Job Scheduler, Hierarchical Timer, Exponential Retries, and Dead-Letter Queue (DLQ) for Go (Go 1.23+).**  
> Schedule delayed actions (e.g. order auto-cancellation after 15m, abandoned cart reminders after 2h, webhook backoff retries) with pluggable persistence and zero external dependencies.

---

## 🚀 Key Features

1. **Precision Timing via Min-Heap Priority Queue**:
   - Jobs are scheduled and popped exactly when their `ExecuteAt` timestamp arrives.
2. **Pluggable Storage Drivers**:
   - `Driver` interface allows seamless plugging of PostgreSQL (with row-level locks / advisory locks), Redis, or SQLite.
   - Includes high-performance, in-memory `MemoryDriver` out-of-the-box.
3. **Automated Exponential Backoff Retries**:
   - When a job handler fails, the engine reschedules with exponential backoff (`RetryBackoff * 2^attempts`).
4. **Dead-Letter Queue (DLQ) Isolation**:
   - Once retries are exhausted, jobs safely transition to `StatusDeadLetter` with root-cause error logs for manual inspection.
5. **Job Lifecycle Cancellation**:
   - Easily cancel scheduled pending jobs before they execute via `q.Cancel(jobID)`.
6. **Zero Dependencies**:
   - 100% pure standard library (`sync`, `container/heap`, `time`, `context`, `encoding/json`).

---

## 📦 Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	"go-delayqueue"
)

func main() {
	// 1. Initialize queue with built-in memory driver
	q := delayqueue.New(delayqueue.NewMemoryDriver())

	// 2. Register topic handler
	q.Register("order.auto_cancel", func(ctx context.Context, job *delayqueue.Job) error {
		fmt.Printf("Auto-cancelling expired order: %s\n", string(job.Payload))
		return nil
	})

	// 3. Start workers & poller
	_ = q.Start()
	defer q.Stop()

	// 4. Schedule job 15 minutes into future
	jobID, _ := q.Enqueue("order.auto_cancel", []byte("ORDER_10924"), 15*time.Minute)
	fmt.Println("Job scheduled with ID:", jobID)
}
```
