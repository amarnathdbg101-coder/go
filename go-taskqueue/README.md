# go-taskqueue ⚡

> **High-Performance Asynchronous Background Worker Pool with Exponential Retries & Panic Recovery.**  
> Built with 100% Go Standard Library, Zero External Dependencies, and Graceful Shutdown.

---

## ⚠️ Why Naked `go doSomething()` is Dangerous in Production

In standard HTTP handlers (e.g. User Signup, Order Checkout), sending emails, calling webhooks, or generating PDFs takes 500ms - 2000ms.

| Issue | Naked `go func()` | `go-taskqueue` (This Library) |
|---|---|---|
| **Server Crash** | 💥 A single panic crashes the entire Go backend | 🛡️ **Panic Recovery**: Isolated per worker; server stays 100% alive |
| **Server Restart / Deploy** | 🗑️ In-flight tasks are instantly killed and lost | 🛑 **Graceful Shutdown**: Waits for active tasks to finish before exit |
| **Spike Traffic (10k requests)** | 💣 Spawns 10,000 unbounded goroutines -> OOM Crash | 🎛️ **Bounded Workers**: Fixed pool (e.g. 8 workers) processes safely |
| **Network Glitches** | ❌ Fails once and gives up | 🔄 **Exponential Backoff Retries**: Auto-retries (500ms, 1s, 2s...) |

---

## ✨ Features

- 🏎️ **Zero External Dependencies**: 100% Pure Go Standard Library (`context`, `sync`, `time`, `encoding/json`).
- 🛡️ **Panic Recovery**: Every task worker runs inside an isolated recovery boundary.
- 🔄 **Exponential Backoff**: Automatic retry scheduling for transient failures (e.g. SMTP/network drops).
- 🛑 **Graceful Shutdown**: `Shutdown(ctx)` ensures zero in-flight task loss during server deployments.
- 🗄️ **Pluggable `Queue` Interface**: High-speed in-memory channel queue built-in; easily adaptable to Redis Streams or SQS.
- 🧬 **Typed JSON Payloads**: `task.UnmarshalPayload(&myStruct)` for clean type-safe handlers.

---

## 📦 Installation

```bash
go get github.com/shopsilo/go-taskqueue
```

---

## ⚡ 2-Minute Quickstart

### Asynchronous Email & Webhook Processing in Chi / net/http

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopsilo/go-taskqueue"
)

type WelcomeEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func main() {
	// 1. Initialize TaskQueue Engine with 8 concurrent workers
	queue := taskqueue.New(
		taskqueue.WithWorkers(8),
		taskqueue.WithBufferSize(1000),
		taskqueue.WithMaxRetries(3),
	)

	// 2. Register Background Task Handlers
	queue.Register("email:welcome", func(ctx context.Context, task *taskqueue.Task) error {
		var payload WelcomeEmailPayload
		if err := task.UnmarshalPayload(&payload); err != nil {
			return err
		}

		log.Printf("📧 [Worker] Sending welcome email to %s (Task ID: %s)...", payload.Email, task.ID)
		time.Sleep(500 * time.Millisecond) // Simulating SMTP send
		log.Printf("✅ [Worker] Email sent to %s!", payload.Email)
		return nil
	})

	// 3. Start Background Workers
	queue.Start()

	// Ensure graceful shutdown on server exit
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = queue.Shutdown(ctx)
	}()

	// 4. HTTP API Handler (Responds in < 1 millisecond!)
	r := chi.NewRouter()
	r.Post("/api/v1/signup", func(w http.ResponseWriter, req *http.Request) {
		// Business logic: save user to PostgreSQL...

		// Dispatch email in background (Non-blocking, takes 0.05ms)
		_, err := queue.Enqueue("email:welcome", WelcomeEmailPayload{
			UserID: "usr_42",
			Email:  "rajesh@example.com",
		})
		if err != nil {
			http.Error(w, "Queue is full", http.StatusServiceUnavailable)
			return
		}

		// User gets instant 201 response without waiting for SMTP!
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"success": true, "message": "User registered"}`))
	})

	log.Println("Server running at :8080")
	http.ListenAndServe(":8080", r)
}
```

---

## 🛡️ Architecture & Concurrency Model

```
Incoming Request (HTTP / Service)
       │
       ▼
queue.Enqueue(taskType, payload)
       │
       ▼
[Bounded Channel Buffer] (e.g. Capacity: 1,000)
       │
       ├── Worker 1 ──► [Panic Recovery] ──► Execute Handler (with 30s Timeout)
       ├── Worker 2 ──► [Panic Recovery] ──► Execute Handler
       ├── Worker ...
       └── Worker N ──► If error ──► Exponential Backoff Retry (500ms -> 1s -> 2s)
```

---

## 🧪 Testing

```bash
cd C:\Library\go-taskqueue
go test -v -count=1 ./...
```
