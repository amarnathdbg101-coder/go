# go-async ⚡

> **Unified Asynchronous Workflows, Worker Pools & Event Engine for Go.**  
> Combines Bounded Worker Pools, Task Queues with Retries, PubSub Fanout, and Event Bus into a single modular Go package.

---

## 📦 Sub-packages

- `github.com/shopsilo/go-async/taskqueue`: Bounded worker pool & panic isolation task queue
- `github.com/shopsilo/go-async/workerpool`: Bounded goroutine pool with backpressure
- `github.com/shopsilo/go-async/events`: In-memory event bus with wildcards & DLQ
- `github.com/shopsilo/go-async/pubsub`: Non-blocking channel fanout

---

## ⚡ 3-Minute Quickstart

```go
package main

import (
	"fmt"
	"time"

	"github.com/shopsilo/go-async/taskqueue"
)

func main() {
	queue, _ := taskqueue.New(taskqueue.Config{Workers: 5})
	queue.Start()
	defer queue.Stop()

	queue.Register("email_welcome", func(t *taskqueue.Task) error {
		fmt.Printf("Sending welcome email to %v\n", t.Payload["email"])
		return nil
	})

	queue.Enqueue(&taskqueue.Task{
		Type:    "email_welcome",
		Payload: map[string]interface{}{"email": "user@example.com"},
	})

	time.Sleep(100 * time.Millisecond)
}
```
