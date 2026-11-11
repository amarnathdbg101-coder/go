# go-distlock 🔐

> [!NOTE]
> **Consolidation Notice**: Yeh package ab full distributed lease capabilities (Monotonic Fencing Tokens, Auto-Heartbeat Lease Extension, aur Pluggable Drivers) ke sath **[`go-locks`](../go-locks)** me consolidate ho chuka hai. Naye projects ke liye unified `go-locks` ka use recommend kiya jata hai. Existing code backward-compatible rahega.

> **Distributed Locking with Monotonically Increasing Fencing Tokens and Auto-Heartbeat Lease Extension for Go (Go 1.23+).**  
> Solves the classic Martin Kleppmann distributed lock anomaly: prevents split-brain writes during GC pauses and network partitions via monotonic fencing tokens and automatic lease renewal.

---

## 🚀 Key Features

1. **Strict Monotonic Fencing Tokens**:
   - Every lock acquisition generates a sequentially increasing 64-bit integer (`uint64`).
   - Downstream databases can enforce `WHERE fencing_token < :token` to reject out-of-order writes from stale/zombie workers.
2. **Auto-Heartbeat Background Renewal**:
   - Dedicated goroutine extends the lock TTL (every $\frac{\text{TTL}}{3}$) as long as the critical section is active, preventing premature lock expiration.
3. **Fail-Fast or Non-Blocking Polling**:
   - Choose immediate failure on contention (`ErrLockNotAcquired`) or configurable retry polling with timeouts (`WithRetry(timeout, interval)`).
4. **Pluggable Storage Drivers**:
   - `Driver` interface accommodates Redis or PostgreSQL advisory locks. Includes an in-memory `MemoryDriver` for zero-dependency usage and local unit testing.
5. **Zero Dependencies**:
   - 100% pure standard library (`sync`, `time`, `context`, `crypto/rand`).

---

## 📦 Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	"go-distlock"
)

func main() {
	ctx := context.Background()

	// 1. Acquire distributed lock with auto-heartbeat
	lock, err := distlock.Acquire(ctx, "inventory:product:99",
		distlock.WithTTL(5*time.Second),
		distlock.WithAutoRenew(true),
	)
	if err != nil {
		fmt.Println("Could not acquire lock (already held):", err)
		return
	}
	defer lock.Release()

	// 2. Obtain fencing token
	fencingToken := lock.FencingToken()
	fmt.Printf("Acquired lock with Fencing Token: %d\n", fencingToken)

	// 3. Perform critical operation safely
	// UPDATE inventory SET quantity = quantity - 1, fencing_token = :fencingToken
	// WHERE id = 99 AND fencing_token < :fencingToken;
}
```
