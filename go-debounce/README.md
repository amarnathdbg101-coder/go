# go-debounce

Thread-safe Debounce & Throttle execution controllers with keyed multi-tenant isolation for Go.

---

## Features
- **Debouncer**: Coalesces rapid bursts into a single execution after a quiet period.
- **Throttler**: Enforces a maximum execution rate (at most once every $N$ milliseconds).
- **Keyed Debouncing**: Independently debounces events per key (e.g. per `userID`, `orderID`, or `documentID`).
- **Cancellation**: Safely aborts pending scheduled timers on shutdown or entity deletion.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

### 1. Simple Debouncer (Autosave)
```go
package main

import (
	"fmt"
	"time"

	godebounce "go-debounce"
)

func main() {
	debouncer := godebounce.New(300 * time.Millisecond)

	// User types in text field rapidly
	for i := 0; i < 5; i++ {
		debouncer.Trigger(func() {
			fmt.Println("Autosaving document to PostgreSQL...")
		})
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)
	// Fires exactly once after the user stops typing!
}
```

### 2. Keyed Debouncer (Multi-User Events)
```go
package main

import (
	"fmt"
	"time"

	godebounce "go-debounce"
)

func main() {
	kd := godebounce.NewKeyed(200 * time.Millisecond)

	// User 1 activity
	kd.Trigger("user_101", func() {
		fmt.Println("Flushing profile for User 101")
	})

	// User 2 activity
	kd.Trigger("user_202", func() {
		fmt.Println("Flushing profile for User 202")
	})

	time.Sleep(300 * time.Millisecond)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
