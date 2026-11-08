# go-semaphore 🚦

`go-semaphore` ek weighted, context-aware counting semaphore library hai Go ke liye jisme FIFO fairness, cancellation refund protection, non-blocking polling (`TryAcquire`), timeouts (`AcquireTimeout`), aur graceful draining features milte hain.

---

## ✨ Features

- **Weighted Tokens**: Multi-token acquisition support (`Acquire(ctx, 3)` ya `Release(3)`).
- **Strict FIFO Fairness**: High load ke dauraan waiters ko fair arrival order me serve karta hai to eliminate starvation.
- **Leak-Proof Context Cancellation**: Context cancel hone par waiter safely queue se detach hota hai aur allocated tokens immediately refund hote hain.
- **Timeouts & Non-Blocking**: `AcquireTimeout(n, 500*time.Millisecond)` aur `TryAcquire(n)`.
- **Drain & Graceful Shutdown**: `Drain()` se active capacity ko safely clear karein aur `Close()` se pending routines ko unblock karein.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-semaphore
```

### Complete Example
```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopsilo/go-semaphore"
)

func main() {
	// Limit max concurrent external API calls to 3
	sem := semaphore.New(3)
	var wg sync.WaitGroup

	for i := 1; i <= 6; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Acquire 1 token with 2-second timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := sem.Acquire(ctx, 1); err != nil {
				fmt.Printf("Worker %d timed out waiting for slot: %v\n", id, err)
				return
			}
			defer sem.Release(1)

			fmt.Printf("Worker %d acquired slot (Available slots: %d)\n", id, sem.Available())
			time.Sleep(100 * time.Millisecond) // simulate workload
		}(i)
	}

	wg.Wait()
	fmt.Println("All tasks processed safely!")
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestSemaphore_BasicAcquireRelease
--- PASS: TestSemaphore_BasicAcquireRelease (0.00s)
=== RUN   TestSemaphore_TryAcquire
--- PASS: TestSemaphore_TryAcquire (0.00s)
=== RUN   TestSemaphore_AcquireTimeout
--- PASS: TestSemaphore_AcquireTimeout (0.07s)
=== RUN   TestSemaphore_ContextCancellation
--- PASS: TestSemaphore_ContextCancellation (0.03s)
=== RUN   TestSemaphore_Drain
--- PASS: TestSemaphore_Drain (0.00s)
=== RUN   TestSemaphore_Close
--- PASS: TestSemaphore_Close (0.03s)
=== RUN   TestSemaphore_ExcessReleaseAndInvalidWeight
--- PASS: TestSemaphore_ExcessReleaseAndInvalidWeight (0.00s)
=== RUN   TestSemaphore_Concurrency
--- PASS: TestSemaphore_Concurrency (0.05s)
PASS
ok  	go-semaphore	0.677s
```

---

## 📄 License
MIT License.
