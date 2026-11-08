# go-concurrentmap

High-throughput, 32-way sharded concurrent hash map with type-safe generics for Go.

---

## Why go-concurrentmap instead of sync.Map?
- **Type-Safe Generics**: No `any` type casting or runtime assertion panics (`Get()` directly returns `V`).
- **Low Mutex Contention**: 32 independent power-of-two shards eliminate lock bottlenecks under high multi-core concurrency.
- **Atomic Operations**: Built-in atomic `GetOrSet`, `Pop`, and callback-driven `Upsert`.
- **Early-Stop Iteration**: Clean `Iter(func(k, v) bool)` traversal.

---

## Quickstart

```go
package main

import (
	"fmt"

	goconcurrentmap "go-concurrentmap"
)

func main() {
	m := goconcurrentmap.New[string, int]()

	// 1. Basic Set & Get
	m.Set("active_users", 450)
	val, _ := m.Get("active_users")
	fmt.Println("Active users:", val)

	// 2. Atomic Upsert (Increment Counter)
	m.Upsert("site_visits", 1, func(exists bool, oldVal, newVal int) int {
		if !exists {
			return newVal
		}
		return oldVal + newVal
	})

	// 3. Safe Iteration
	m.Iter(func(key string, val int) bool {
		fmt.Printf("- %s: %d\n", key, val)
		return true // Continue iteration
	})
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
