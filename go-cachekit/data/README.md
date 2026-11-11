# go-cachekit/data 🗄️

> Cache facade and multi-tier caching packages within the consolidated `go-cachekit` module.

---

## 📦 Sub-packages

- `github.com/amarnathdbg101-coder/go/go-cachekit/data`: Small facade exposing the cache engine and store types
- `github.com/amarnathdbg101-coder/go/go-cachekit/data/cache`: L1 RAM + L2 Redis, singleflight protection, and typed `GetOrSet[T]`
- `github.com/amarnathdbg101-coder/go/go-cachekit/lru`, `/lfu`, and `/tinylfu`: Separate eviction-policy packages in the same module

---

## ⚡ 3-Minute Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/amarnathdbg101-coder/go/go-cachekit/data/cache"
)

type Product struct {
	ID    string  `json:"id"`
	Price float64 `json:"price"`
}

func main() {
	store := cache.NewMemoryStore()
	defer store.Close()

	engine, _ := cache.New(store)
	prodCache := cache.Typed[Product](engine)

	prod, _ := prodCache.GetOrSet(context.Background(), "prod:101", 10*time.Minute, func(ctx context.Context) (Product, error) {
		return Product{ID: "101", Price: 1999.0}, nil
	})

	fmt.Printf("Product: %+v\n", prod)
}
```
