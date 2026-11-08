# go-ringbuffer

Thread-safe, generic circular FIFO ring buffer with overwrite and bounded modes for Go.

---

## Features
- **Generic Circular Buffer**: Type-safe storage for any item type `T`.
- **Zero Reallocations**: Operates strictly within pre-allocated capacity.
- **Configurable Overflow Strategy**:
  - `overwrite = true`: Automatically drops oldest items (ideal for sliding logs and real-time streaming).
  - `overwrite = false`: Rejects overflow with `ErrBufferFull` (ideal for bounded queues).
- **GC Safe**: Zeroes out array slots upon `Pop` to prevent memory leaks of large objects.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	goringbuffer "go-ringbuffer"
)

func main() {
	// Circular log tail keeping only last 3 log messages
	buf, _ := goringbuffer.New[string](3, true)

	_ = buf.Push("log 1: startup")
	_ = buf.Push("log 2: connected")
	_ = buf.Push("log 3: user joined")
	_ = buf.Push("log 4: query executed") // drops "log 1: startup"

	// Inspect current sliding window
	for _, msg := range buf.Slice() {
		fmt.Println("Active:", msg)
	}
	// Output:
	// Active: log 2: connected
	// Active: log 3: user joined
	// Active: log 4: query executed
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
