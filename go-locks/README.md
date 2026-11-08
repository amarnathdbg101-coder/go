# go-locks 🔒

Distributed & in-memory locking primitives: zero-leak reference-counted KeyedMutex and PostgreSQL advisory locks for race condition prevention. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🗝️ **KeyedMutex (In-Memory)**: Granular per-key locking (`user_102` vs `user_103`). Alag-alag users parallel run karte hain, jabki same user ki concurrent requests serialize ho jati hain.
- 🚯 **Zero-Leak Auto-Cleanup**: Reference counting ke zariye jaise hi key ka lock unlock hota hai, map se mutex entry delete ho jati hai. 10 million unique keys par bhi 0 RAM leak!
- 🐘 **PostgreSQL Advisory Locks**: Distributed multi-instance locking bina kisi Redis dependency ke directly PostgreSQL ke `pg_try_advisory_lock` se.
- ⚡ **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. In-Memory Keyed Locking (Preventing Double Spending)
```go
package main

import (
	"fmt"
	golocks "go-locks"
)

var userLocks = golocks.NewKeyedMutex()

func debitWallet(userID string, amount float64) error {
	// Only lock this specific user; other users continue concurrently!
	return userLocks.WithLock(userID, func() error {
		fmt.Printf("Processing wallet debit for user %s: ₹%.2f\n", userID, amount)
		// Database select, balance check, and debit
		return nil
	})
}
```

### 2. Distributed Locking with PostgreSQL
```go
import (
	"context"
	"database/sql"
	"log"
	"time"
	golocks "go-locks"
)

func runNightlyReconciliation(db *sql.DB) {
	ctx := context.Background()
	// Distributed lock across all k8s replicas
	locker := golocks.NewPostgresLockerFromString("cron:nightly_reconciliation")

	acquired, err := locker.TryLock(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	if !acquired {
		log.Println("Another instance is already running this cron job. Skipping.")
		return
	}
	defer locker.Unlock(ctx, db)

	log.Println("Acquired cluster lock! Running reconciliation...")
	time.Sleep(5 * time.Second)
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
