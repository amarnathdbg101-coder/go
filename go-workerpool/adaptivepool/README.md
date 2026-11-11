# go-workerpool/adaptivepool 🌊

> 💡 **Consolidation Notice**: Dynamic elastic auto-scaling and idle worker eviction have been consolidated directly into [`go-workerpool`](../go-workerpool) via `goworkerpool.NewAdaptive(...)`. For new projects, prefer importing [`go-workerpool`](../go-workerpool) directly for both static bounded and elastic auto-scaling pools in a single package.

> **Auto-Scaling Elastic Worker Pool with Idle Eviction for Go (Go 1.23+).**  
> Static worker pools ya toh load aane par tasks queue me accumulate karte hain ya fir shuru se bohot saari idle goroutines chala kar RAM waste karte hain. `go-workerpool/adaptivepool` baseline workers (`MinWorkers`) se start hota hai, load aane par dynamically `MaxWorkers` tak scale karta hai, aur idle hone par surplus goroutines ko automatically terminate karke RAM free karta hai.

---

## 🚀 Features

- **Elastic Scaling**: Normal time par `MinWorkers` chalte hain; queue me load aate hi `MaxWorkers` tak expand hote hain.
- **Idle Worker Eviction**: Jo extra workers spawn huye the, unhe `IdleTimeout` ke baad safely terminate kiya jata hai.
- **Backpressure Submissions**:
  - `Submit`: Blocking queue push.
  - `TrySubmit`: Non-blocking instant enqueue.
  - `SubmitWithTimeout`: Bounded wait time.
- **Panic Isolation**: Task me panic aane par goroutine pool crash nahi hota.
- **Graceful Drain**: `Close()` call karne par bache huye queue tasks complete hote hain fir sabhi workers exit karte hain.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"github.com/amarnathdbg101-coder/go/go-workerpool/adaptivepool"
	"time"
)

func main() {
	// Min 2 workers, Max 20 workers, 2s idle timeout
	pool := adaptivepool.New(adaptivepool.Config{
		MinWorkers:  2,
		MaxWorkers:  20,
		IdleTimeout: 2 * time.Second,
		QueueSize:   100,
	})
	defer pool.Close()

	// Tasks submit karein
	for i := 1; i <= 50; i++ {
		taskID := i
		_ = pool.Submit(func() {
			fmt.Printf("Executing Task #%d (Active Workers: %d)\n", taskID, pool.ActiveWorkers())
			time.Sleep(50 * time.Millisecond)
		})
	}

	time.Sleep(1 * time.Second)
}
```
