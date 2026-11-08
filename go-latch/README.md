# go-latch ⏱️

`go-latch` ek advanced synchronization primitives library hai Go ke liye jo `CountDownLatch` aur `CyclicBarrier` provide karti hai. Ye standard `sync.WaitGroup` se aage badh kar timed waiting, context cancellation, multi-stage pipeline barrier coordination, aur broken barrier error propagation support karti hai.

---

## ✨ Features

- **CountDownLatch**:
  - `WaitTimeout(duration)`: Blocks until latch hits 0 ya timeout elapse ho jaye.
  - `WaitContext(ctx)`: Graceful cancellation with context propagation.
  - Atomic zero count closure with zero deadlocks.
- **CyclicBarrier**:
  - $N$ parallel workers ko cyclic checkpoints par synchronize karta hai.
  - Automatic barrier action hook trigger on the last arrival.
  - Automatically resets for next iteration cycle.
  - Broken barrier detection agar koi ek worker cancel ya timeout ho jaye.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-latch
```

### CountDownLatch Example with Timeout
```go
package main

import (
	"fmt"
	"time"

	"github.com/shopsilo/go-latch"
)

func main() {
	latch := latch.NewCountDownLatch(2)

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Worker 1 finished DB migration")
		latch.CountDown()
	}()

	go func() {
		time.Sleep(150 * time.Millisecond)
		fmt.Println("Worker 2 finished Cache warmup")
		latch.CountDown()
	}()

	// Wait up to 1 second
	if latch.WaitTimeout(1 * time.Second) {
		fmt.Println("All startup dependencies ready! Starting HTTP server...")
	} else {
		fmt.Println("Startup timed out!")
	}
}
```

### CyclicBarrier Multi-Stage Example
```go
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/shopsilo/go-latch"
)

func main() {
	const workers = 3
	barrier := latch.NewCyclicBarrier(workers, func() {
		fmt.Println("--- Stage Complete! Synchronizing to next stage ---")
	})

	var wg sync.WaitGroup
	for id := 1; id <= workers; id++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Stage 1
			fmt.Printf("Worker %d finished Stage 1\n", workerID)
			barrier.Await(context.Background())

			// Stage 2
			fmt.Printf("Worker %d finished Stage 2\n", workerID)
			barrier.Await(context.Background())
		}(id)
	}
	wg.Wait()
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestCountDownLatch_Basic
--- PASS: TestCountDownLatch_Basic (0.02s)
=== RUN   TestCountDownLatch_Timeout
--- PASS: TestCountDownLatch_Timeout (0.03s)
=== RUN   TestCountDownLatch_Context
--- PASS: TestCountDownLatch_Context (0.02s)
=== RUN   TestCyclicBarrier_MultiCycle
--- PASS: TestCyclicBarrier_MultiCycle (0.02s)
=== RUN   TestCyclicBarrier_BrokenOnCancel
--- PASS: TestCyclicBarrier_BrokenOnCancel (0.03s)
PASS
ok  	go-latch	0.606s
```

---

## 📄 License
MIT License.
