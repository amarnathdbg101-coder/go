# go-pool

Bounded, generic resource pool with active health checking and idle timeout eviction for Go.

---

## Why go-pool instead of sync.Pool?
| Feature | `sync.Pool` | `go-pool` |
| :--- | :--- | :--- |
| **Capacity Bounded** | ❌ Unbounded (Allocates infinitely) | ✅ Strictly Bounded (Max $N$ items) |
| **GC Resilience** | ❌ Clears all items every GC cycle | ✅ Retains long-lived items |
| **Health Checks** | ❌ No health validation | ✅ Active `HealthCheck` before use |
| **Idle Eviction** | ❌ No idle timeout | ✅ Discards stale connections automatically |
| **Context Aware** | ❌ Cannot cancel or timeout | ✅ Honors `context.Context` timeout |

---

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"net"
	"time"

	gopool "go-pool"
)

func main() {
	pool, err := gopool.New(gopool.Config[net.Conn]{
		Capacity:    10,                       // Maximum 10 connections
		IdleTimeout: 1 * time.Minute,          // Close if idle for >1m
		Factory: func(ctx context.Context) (net.Conn, error) {
			return net.Dial("tcp", "127.0.0.1:5432")
		},
		CloseFn: func(c net.Conn) error {
			return c.Close()
		},
	})
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// Acquire connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		fmt.Println("Pool acquire failed:", err)
		return
	}

	// Use connection...
	fmt.Println("Connection acquired successfully!")

	// Return connection to pool for reuse
	pool.Release(conn)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
