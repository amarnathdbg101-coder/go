# go-cron

Lightweight, zero-dependency, panic-isolated in-memory cron and periodic job scheduler for Go applications.

---

## Features
- **Standard Unix 5-Field Cron**: Supports `"minute hour day-of-month month day-of-week"` syntax with wildcards (`*`), steps (`*/5`), ranges (`1-5`), and lists (`1,15,30`).
- **Interval Scheduling**: Simple, ergonomic interval registration via `Every(10 * time.Second)`.
- **Panic Isolation**: Individual job panics are caught and reported without crashing the scheduler loop or host application.
- **Asynchronous Execution**: Every job runs concurrently in its own goroutine so slow tasks never delay subsequent scheduled executions.
- **Graceful Lifecycle Management**: Safe `Start()`, `Stop()`, and dynamic task removal by ID.

---

## Installation
```bash
go get go-cron
```
*(Or include directly as a local module)*

---

## Quickstart

```go
package main

import (
	"fmt"
	"time"

	gocron "go-cron"
)

func main() {
	s := gocron.New()

	// Panic isolation handler
	s.SetOnPanic(func(jobName string, r any) {
		fmt.Printf("[CRON ERROR] Job %s panicked: %v\n", jobName, r)
	})

	// 1. Run every 500 milliseconds
	s.Every("heartbeat", 500*time.Millisecond, func() {
		fmt.Println("Heartbeat tick:", time.Now().Format("15:04:05.000"))
	})

	// 2. Standard 5-field cron: Midnight cleanup
	_, err := s.AddCron("midnight-cleanup", "0 0 * * *", func() {
		fmt.Println("Running midnight database maintenance...")
	})
	if err != nil {
		panic(err)
	}

	// 3. Start scheduler
	if err := s.Start(); err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)

	// Graceful shutdown
	s.Stop()
}
```

---

## Cron Syntax Cheatsheet
| Expression | Description |
| :--- | :--- |
| `* * * * *` | Run every minute |
| `*/5 * * * *` | Run every 5 minutes |
| `0 * * * *` | Run every hour at minute 0 |
| `30 2 * * *` | Run daily at 02:30 AM |
| `0 0 * * 0` | Run every Sunday at midnight |
| `0 9 1-5 * *` | Run at 09:00 AM on days 1 through 5 of each month |

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
