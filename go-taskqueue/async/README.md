# go-taskqueue/async ⚡

> Asynchronous task queue package within the consolidated `go-taskqueue` module.

---

## 📦 Sub-packages

- `github.com/amarnathdbg101-coder/go/go-taskqueue/async`: Async runner and lifecycle
- `github.com/amarnathdbg101-coder/go/go-taskqueue/async/taskqueue`: Bounded task queue with retries and panic isolation

---

## ⚡ 3-Minute Quickstart

```go
package main

import (
	"fmt"
	"time"

	"github.com/amarnathdbg101-coder/go/go-taskqueue/async/taskqueue"
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
