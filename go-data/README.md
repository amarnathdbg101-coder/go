# go-data 🗄️

> **Unified Data Caching, SQL Building, Pagination & Validation Suite for Go.**  
> Combines Multi-tier Caching, Singleflight, Cursor Pagination, SQL Builder, and Fluent Validator into a single modular Go package.

---

## 📦 Sub-packages

- `github.com/shopsilo/go-data/cache`: L1 RAM + L2 Redis, Singleflight Dogpile protection, Generics `GetOrSet[T]`
- `github.com/shopsilo/go-data/paginate`: SQL Injection-Proof Cursor & Offset Pagination
- `github.com/shopsilo/go-data/sqlbuilder`: Parameterized SQL Query Builder
- `github.com/shopsilo/go-data/validator`: Reflection-free fluent validator & tag rules
- `github.com/shopsilo/go-data/sanitize`: XSS & SQL LIKE sanitizer

---

## ⚡ 3-Minute Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/shopsilo/go-data/cache"
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
