# go-crdt 🔄

> **Conflict-Free Replicated Data Types (PNCounter, LWWRegister, ORSet) for Go (Go 1.23+).**  
> Enables offline-first mobile sync (e.g. Flutter POS & distributed microservices) to converge deterministically to identical states without centralized locks or merge conflicts.

---

## 🚀 Key Features

1. **State-Based CRDTs (CvRDT)**:
   - All merges form a mathematically proven **Join-Semilattice**:
     - **Associative**: $(A \vee B) \vee C = A \vee (B \vee C)$
     - **Commutative**: $A \vee B = B \vee A$
     - **Idempotent**: $A \vee A = A$
2. **`PNCounter`**:
   - Supports concurrent increments and decrements across disconnected devices. Merges calculate $\max(P)$ and $\max(N)$ per node.
3. **`LWWRegister[T any]`**:
   - Generic Last-Write-Wins Register with nanosecond timestamp tracking and deterministic Node ID tie-breaking.
4. **`ORSet[T comparable]`**:
   - Observed-Remove Set (Add-Wins Set) with unique observation tags, preventing tombstone resurrection anomalies.
5. **Zero Dependencies**:
   - 100% pure standard library (`sync`, `time`, `encoding/json`, `crypto/rand`).

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-crdt"
)

func main() {
	// Cashier 1 (Terminal A) sells 5 items offline
	pos1 := crdt.NewPNCounter("counter_1")
	pos1.Inc(5)

	// Cashier 2 (Terminal B) accepts 2 item returns offline
	pos2 := crdt.NewPNCounter("counter_2")
	pos2.Dec(2)

	// When internet reconnects, merge both states:
	_ = pos1.Merge(pos2)
	_ = pos2.Merge(pos1)

	// Both terminals converge to identical value: 3!
	fmt.Printf("Terminal 1 Stock Change: %d\n", pos1.Value()) // 3
	fmt.Printf("Terminal 2 Stock Change: %d\n", pos2.Value()) // 3
}
```
