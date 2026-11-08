# go-atomic ⚛️

`go-atomic` ek ergonomic, type-safe suite of high-level atomic primitives library hai Go ke liye. Ye raw `sync/atomic` ke low-level pointer boilerplate ko replace karta hai with lock-free wrappers for `Bool` (with atomic `.Toggle()`), `String`, `Duration`, `Time`, `Float64` (with CAS math), aur generic `Value[T]`.

---

## ✨ Features

- **Ergonomic Atomic Bool**: `.Load()`, `.Store()`, `.Swap()`, `.CAS()`, aur atomic `.Toggle()`.
- **Race-Free Strings**: Safely update string configuration without data races (`String.Store()` / `Load()`).
- **Atomic Duration & Timestamp**: Thread-safe latency benchmarks and clock tracking (`Duration.Add()`, `Time.Store()`).
- **Atomic Float64**: IEEE-754 bitwise CAS loop for thread-safe pricing, rates, and counters (`Float64.Add()`).
- **Type-Safe Generic `Value[T]`**: Safely swap complex structs without interface casting panics.
- **JSON Serialization**: Direct marshaling support for configs and metric dumps.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-atomic
```

### Complete Example
```go
package main

import (
	"fmt"
	"time"

	"github.com/shopsilo/go-atomic"
)

type ServerState struct {
	ActiveConnections atomic.Float64
	IsDraining        *atomic.Bool
	LastHealthCheck   *atomic.Time
	CurrentVersion    *atomic.String
}

func main() {
	state := ServerState{
		IsDraining:      atomic.NewBool(false),
		LastHealthCheck: atomic.NewTime(time.Now()),
		CurrentVersion:  atomic.NewString("v1.4.2"),
	}

	// Atomic toggle
	isDrainingNow := state.IsDraining.Toggle()
	fmt.Printf("Server draining mode toggled: %v\n", isDrainingNow) // true

	// Atomic float increment
	state.ActiveConnections.Add(1.0)
	fmt.Printf("Active connections: %.0f\n", state.ActiveConnections.Load()) // 1

	// Atomic string hot-reload
	oldVer := state.CurrentVersion.Swap("v1.5.0")
	fmt.Printf("Upgraded from %s to %s\n", oldVer, state.CurrentVersion.Load())
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestBool
--- PASS: TestBool (0.00s)
=== RUN   TestString
--- PASS: TestString (0.00s)
=== RUN   TestDuration
--- PASS: TestDuration (0.00s)
=== RUN   TestTime
--- PASS: TestTime (0.00s)
=== RUN   TestFloat64
--- PASS: TestFloat64 (0.00s)
=== RUN   TestValueGeneric
--- PASS: TestValueGeneric (0.00s)
=== RUN   TestConcurrency
--- PASS: TestConcurrency (0.00s)
PASS
ok  	go-atomic	0.537s
```

---

## 📄 License
MIT License.
