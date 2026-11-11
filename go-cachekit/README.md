# go-cachekit ⚡

> **High-Performance Multi-Tier Caching with Cache Stampede Protection and Go Generics.**  
> Built with 100% Go Standard Library idioms, Zero External Dependencies, and Microsecond Latency.

---

## ✨ Features

- 🏎️ **Zero External Dependencies**: 100% Pure Go Standard Library (`sync`, `time`, `context`, `encoding/json`).
- 🛡️ **Cache Stampede / Dogpile Protection**: Built-in **Singleflight** request coalescing. 10,000 concurrent requests result in **exactly 1 database query**.
- 🧬 **Go Generics (`TypedCache[T]`)**: Type-safe caching for any struct without manual `json.Marshal` or `interface{}` type casting.
- 🚀 **Multi-Tier Architecture (TwoTierStore)**:
  - **L1 In-Memory (RAM)**: Microsecond access (~1-3 µs).
  - **L2 Distributed (Redis/Remote)**: Millisecond access (~1-2 ms) with auto-backfilling of L1 on misses.
- 🧹 **Passive & Active Eviction**: Passive expiration on read + background ticker sweeping to prevent memory leaks.
- 🔒 **Concurrency-Safe**: Built with `sync.RWMutex` readers-writer locks for non-blocking concurrent reads.

---

## 📦 Installation

```bash
go get github.com/amarnathdbg101-coder/go/go-cachekit
```

---

## ⚡ 2-Minute Quickstart

### 1. In-Memory Caching with PostgreSQL Database Fallback

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/amarnathdbg101-coder/go/go-cachekit"
)

type Product struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

func main() {
	ctx := context.Background()

	// 1. Initialize In-Memory Store & Cache Engine
	store := cachekit.NewMemoryStore()
	defer store.Close()

	engine, err := cachekit.New(store)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Create strongly-typed Product cache
	productCache := cachekit.Typed[Product](engine)

	// 3. One-line Read-Through Cache with Dogpile Protection!
	// If in cache -> returns in 2 microseconds
	// If miss -> queries PostgreSQL once, caches for 10 minutes, returns to caller
	product, err := productCache.GetOrSet(ctx, "prod:101", 10*time.Minute, func(ctx context.Context) (Product, error) {
		log.Println("⚡ Cache miss: Fetching from PostgreSQL...")
		return fetchProductFromPostgres(ctx, "101")
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Fetched Product: %+v\n", product)
}

func fetchProductFromPostgres(ctx context.Context, id string) (Product, error) {
	// Simulating DB call: SELECT id, title, price FROM products WHERE id = $1
	return Product{ID: id, Title: "MacBook Pro M3", Price: 199999.00}, nil
}
```

---

## 🛡️ Why Cache Stampede (Dogpile) Protection Matters

### Without Singleflight (Standard Cache):
```
10,000 Users Request /products/hot at 12:00:00 (Cache expires)
       │
       ├── Request 1  ───► DB Query 1
       ├── Request 2  ───► DB Query 2
       ├── Request ... ──► DB Query ...
       └── Request 10k ──► DB Query 10,000! 💥 (Database CPU Spikes to 100%, crashes)
```

### With `go-cachekit` Singleflight:
```
10,000 Users Request /products/hot at 12:00:00 (Cache expires)
       │
       ├── Request 1   ───► Executes 1 Database Query (Takes 40ms)
       └── Requests 2..10k wait and coalesce onto Request 1's channel
                           │
                           ▼
       All 10,000 Users receive the result simultaneously! 
       ⚡ Database Load: EXACTLY 1 QUERY.
```

---

## 🚀 Two-Tier Hybrid Caching (L1 RAM + L2 Redis)

For multi-instance distributed deployments:

```go
package main

import (
	"context"
	"time"

	"github.com/amarnathdbg101-coder/go/go-cachekit"
)

func setupTwoTierCache(redisStore cachekit.Store) *cachekit.Engine {
	// L1: Local RAM (Microsecond response)
	l1 := cachekit.NewMemoryStore()

	// L2: Shared Redis/Remote Store (Millisecond response)
	l2 := redisStore

	// TwoTier: L1 miss -> checks L2 -> auto-backfills L1 -> returns
	twoTier := cachekit.NewTwoTierStore(l1, l2)

	engine, _ := cachekit.New(twoTier)
	return engine
}
```

---

## 🧹 Cache Invalidation Patterns

```go
// 1. Evict single key on update
productCache.Delete(ctx, "prod:101")

// 2. Invalidate an entire category namespace (e.g. on new product creation)
engine.DeletePrefix(ctx, "category:electronics:")

// 3. Flush everything
engine.Clear(ctx)
```

---

## 🧪 Testing

```bash
cd C:\Library\go-cachekit
go test -v -count=1 ./...
```
