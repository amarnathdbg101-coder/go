# go-batcher

High-throughput, generic, thread-safe micro-batching aggregator for Go applications.

---

## Features
- **Dual-Trigger Flushes**: Automatically flushes when count reaches `MaxBatchSize` or duration exceeds `FlushInterval`.
- **Zero Item Loss on Shutdown**: `Stop()` guarantees all buffered and pending queue items are flushed before termination.
- **Backpressure Protection**: Supports both blocking `Push(item)` and non-blocking `TryPush(item)` returning `ErrQueueFull`.
- **Panic Isolated**: Flush panics are safely caught and redirected to `OnPanic` without crashing the scheduler loop.
- **Type-Safe Generics**: Works with any item type `T` without allocations or type casting.

---

## Quickstart

```go
package main

import (
	"fmt"
	"time"

	gobatcher "go-batcher"
)

type AuditEvent struct {
	UserID string
	Action string
}

func main() {
	batcher, err := gobatcher.New(gobatcher.Config[AuditEvent]{
		MaxBatchSize:  500,               // Flush every 500 items
		FlushInterval: 100 * time.Millisecond, // Or every 100ms
		FlushFn: func(batch []AuditEvent) error {
			fmt.Printf("Flushing batch of %d events to database...\n", len(batch))
			return nil
		},
	})
	if err != nil {
		panic(err)
	}
	defer batcher.Stop()

	// Push items concurrently
	for i := 0; i < 1250; i++ {
		_ = batcher.Push(AuditEvent{
			UserID: fmt.Sprintf("usr_%d", i),
			Action: "LOGIN",
		})
	}

	// Manual immediate flush
	batcher.Flush()
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
