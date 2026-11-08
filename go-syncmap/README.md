# go-syncmap 🗺️

`go-syncmap` ek high-performance, type-safe generic wrapper hai Go ke standard `sync.Map` ke liye (`Map[K comparable, V any]`). Ye runtime `any` type assertion panics ko compile-time safety ke sath replace karta hai aur atomic `Compute` transactions, `Keys()`, `Values()`, `Len()`, aur `Has()` utilities provide karta hai.

---

## ✨ Features

- **Compile-Time Type Safety**: No more `val.(string)` or `k.(int)` casting bugs.
- **Full Standard Parity**: `Load`, `Store`, `LoadOrStore`, `LoadAndDelete`, `Delete`, `Swap`, `CompareAndSwap`, `CompareAndDelete`.
- **Atomic Read-Modify-Write**: `Compute(key, fn)` se thread-safe counters ya cache refresh karein without manual mutexes.
- **Inspection Utilities**: `Keys()`, `Values()`, `Len()`, `Has()`, `Clone()`, `Equal()`.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-syncmap
```

### Complete Example
```go
package main

import (
	"fmt"

	"github.com/shopsilo/go-syncmap"
)

type Session struct {
	UserID string
	IP     string
}

func main() {
	// Strongly-typed concurrent map
	sessions := syncmap.New[string, Session]()

	// 1. Store & Load
	sessions.Store("sess-1", Session{UserID: "usr-100", IP: "192.168.1.1"})

	sess, ok := sessions.Load("sess-1")
	if ok {
		fmt.Printf("User: %s from IP: %s\n", sess.UserID, sess.IP)
	}

	// 2. Atomic Compute (e.g. Rate-limit counter)
	hits := syncmap.New[string, int]()
	hits.Compute("192.168.1.1", func(oldVal int, exists bool) (int, bool) {
		return oldVal + 1, false // false means keep in map
	})

	count, _ := hits.Load("192.168.1.1")
	fmt.Printf("Hits count: %d\n", count) // 1

	// 3. Inspect keys
	fmt.Printf("Active session keys: %v (Total: %d)\n", sessions.Keys(), sessions.Len())
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestMap_BasicOps
--- PASS: TestMap_BasicOps (0.00s)
=== RUN   TestMap_LoadOrStoreAndSwap
--- PASS: TestMap_LoadOrStoreAndSwap (0.00s)
=== RUN   TestMap_CompareOperations
--- PASS: TestMap_CompareOperations (0.00s)
=== RUN   TestMap_KeysValuesRange
--- PASS: TestMap_KeysValuesRange (0.00s)
=== RUN   TestMap_Compute
--- PASS: TestMap_Compute (0.00s)
=== RUN   TestMap_Concurrency
--- PASS: TestMap_Concurrency (0.00s)
PASS
ok  	go-syncmap	0.485s
```

---

## 📄 License
MIT License.
