# go-sqlbuilder 🏗️

Zero-reflection, lightweight SQL-injection safe query builder for PostgreSQL (`$1, $2, ...`) without heavy ORM overhead. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🛡️ **SQL Injection Immunity**: Query values ko kabhi SQL string me concatenate nahi karta; strictly parameterized arguments (`$1, $2, ...`) generate karta hai.
- ⚡ **Zero Reflection**: Extreme performance, no struct tag parsing or hidden ORM N+1 queries.
- 🧩 **Dynamic Filtering**: Optional filters (`Where`, `WhereIn`, `Join`, `OrderBy`, `Limit`, `Offset`) ko easily dynamically compose karta hai.
- 🐘 **PostgreSQL Native**: Built-in support for `RETURNING` clauses and `$1` parameter syntax.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Dynamic SELECT with Filtering
```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	gosql "go-sqlbuilder"
)

func searchProducts(db *sql.DB, category string, minPrice float64) {
	b := gosql.Select("id", "title", "price").
		From("products").
		Where("status = ?", "active")

	if category != "" {
		b.Where("category = ?", category)
	}
	if minPrice > 0 {
		b.Where("price >= ?", minPrice)
	}

	query, args, err := b.OrderBy("price", "ASC").Limit(10).Build()
	if err != nil {
		log.Fatal(err)
	}

	// Directly pass to database/sql
	rows, err := db.Query(query, args...)
	_ = rows
}
```

### 2. INSERT with RETURNING
```go
query, args, _ := gosql.InsertInto("orders").
    Columns("user_id", "total_amount", "status").
    Values(42, 2499.00, "paid").
    Returning("id", "created_at").
    Build()

// query: INSERT INTO orders (user_id, total_amount, status) VALUES ($1, $2, $3) RETURNING id, created_at
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
