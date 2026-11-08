# go-kvstore

High-concurrency, sharded in-memory key-value cache with TTL expiration and atomic operations for Go.

---

## Features
- **32-Way Sharded Lock Striping**: Minimizes lock contention under massive multi-threaded load.
- **Dual TTL Expiration**: Combines lazy eviction on `Get()` with periodic active background sweeping.
- **Atomic Operations**: Out-of-the-box `GetOrSet`, `CompareAndSwap` (CAS), and integer `Increment`.
- **Sub-Microsecond Lookups**: High-speed memory operations with zero external services (no Redis required for single-node caching).
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"
	"time"

	gokvstore "go-kvstore"
)

func main() {
	store := gokvstore.New(gokvstore.DefaultConfig())
	defer store.Close()

	// 1. Set key with 2-second TTL
	store.Set("session:abc123", "user_profile_data", 2*time.Second)

	val, found := store.Get("session:abc123")
	if found {
		fmt.Println("Found session:", val)
	}

	// 2. Atomic counter increment
	count, _ := store.Increment("site_visits", 1)
	fmt.Printf("Current visits: %d\n", count)

	// 3. Compare and Swap (Atomic state transition)
	store.Set("deploy_state", "READY", 0)
	if store.CompareAndSwap("deploy_state", "READY", "DEPLOYING") {
		fmt.Println("Successfully transitioned deploy state!")
	}
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
