# go-mvcc

`go-mvcc` provides a production-grade Multi-Version Concurrency Control (MVCC) engine in pure Go. It delivers true Snapshot Isolation (SI) and Repeatable Reads—readers never block writers, and writers never block readers—mirroring the transaction core of PostgreSQL, CockroachDB, and TiDB.

## Key Features

- **Snapshot Isolation (SI)**: Every transaction operates against an immutable point-in-time database snapshot determined at `Begin()`.
- **First-Committer-Wins Conflict Detection**: Prevents write-write lost updates by immediately aborting conflicting concurrent commits with `ErrWriteConflict`.
- **Version Chains**: Non-destructive updates create version records maintaining audit lineage.
- **Vacuuming & Garbage Collection**: Automatically prunes superseded dead tuples older than any active transaction snapshot.
- **Rollback Safety**: Aborted transactions cleanly discard staged mutations without leaving uncommitted remnants.
- **Zero Dependencies**: Pure Go standard library (`sync`, `sync/atomic`).

## Quick Start

```go
package main

import (
	"fmt"
	"go-mvcc"
)

func main() {
	engine := mvcc.New()

	// Initial commit
	tx0 := engine.Begin()
	_ = tx0.Set([]byte("balance"), []byte("5000"))
	_ = tx0.Commit()

	// Tx1 starts snapshot
	tx1 := engine.Begin()

	// Tx2 concurrently updates balance
	tx2 := engine.Begin()
	_ = tx2.Set([]byte("balance"), []byte("9000"))
	_ = tx2.Commit()

	// Readers never block writers: Tx1 still sees 5000!
	val1, _ := tx1.Get([]byte("balance"))
	fmt.Printf("Tx1 Snapshot Isolation balance: %s\n", string(val1)) // 5000
	_ = tx1.Commit()

	// New transaction sees updated 9000
	tx3 := engine.Begin()
	val3, _ := tx3.Get([]byte("balance"))
	fmt.Printf("Tx3 Current committed balance: %s\n", string(val3)) // 9000
	_ = tx3.Commit()
}
```
