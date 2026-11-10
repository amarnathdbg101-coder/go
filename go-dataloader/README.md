# go-dataloader ⚡

> **High-Performance, Zero-Dependency Generic Batching and Caching Data Loader for Go (Go 1.23+).**  
> Coalesces concurrent single-key lookups into a single batched database or microservice call, completely eliminating the $N+1$ query problem in REST and GraphQL services.

---

## 🚀 Key Features

1. **Automatic Request Coalescing**:
   - Collects all requests arriving within a configurable debounce window (e.g. 2ms) and passes them to your batch function together (`SELECT * FROM users WHERE id IN ($1, $2, ...)`).
2. **Go 1.23+ Type-Safe Generics (`[K comparable, V any]`)**:
   - Zero reflection (`reflect.ValueOf`) overhead.
3. **Per-Request Memoization Cache**:
   - Avoids duplicate database fetches within the same lifecycle. Includes `Prime(key, val)`, `Clear(key)`, and `ClearAll()`.
4. **Max Batch Early Dispatch**:
   - Automatically flushes immediately when `MaxBatch` limit is hit, avoiding waiting for the full timeout.
5. **Context-Aware**:
   - Respects `context.Context` timeouts and cancellations without leaking goroutines or channels.
6. **Zero Dependencies**:
   - 100% pure Go standard library (`sync`, `time`, `context`).

---

## 📦 Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	"go-dataloader"
)

type User struct {
	ID   int
	Name string
}

func main() {
	// 1. Define batch loading function (e.g. PostgreSQL SELECT ... WHERE id IN (...))
	batchUsers := func(ctx context.Context, ids []int) ([]dataloader.Result[User], error) {
		fmt.Printf("--> Database queried once for %d user IDs: %v\n", len(ids), ids)
		results := make([]dataloader.Result[User], len(ids))
		for i, id := range ids {
			results[i] = dataloader.Result[User]{
				Data: User{ID: id, Name: fmt.Sprintf("User_%d", id)},
			}
		}
		return results, nil
	}

	// 2. Initialize loader
	loader := dataloader.New(batchUsers, dataloader.WithWait(2*time.Millisecond))
	defer loader.Close()

	ctx := context.Background()

	// 3. Concurrent requests from different handlers
	go func() {
		u, _ := loader.Load(ctx, 101)
		fmt.Println("Handler 1 received:", u.Name)
	}()

	go func() {
		u, _ := loader.Load(ctx, 102)
		fmt.Println("Handler 2 received:", u.Name)
	}()

	time.Sleep(20 * time.Millisecond)
}
```
