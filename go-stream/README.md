# go-stream 🌊

Type-safe, zero-dependency functional stream pipeline and concurrent batch processor for Go Generics `[T any]`. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🧬 **Go Generics**: Type-safe functional pipeline (`Filter`, `Map`, `Reduce`, `Distinct`, `Sort`, `Chunk`, `FlatMap`) without reflection.
- ⚡ **Order-Preserving `ParallelMap`**: Elements ko background goroutines me parallel process karta hai jabki output ka order bilkul intact rehta hai.
- 📦 **Batch Chunking**: Array ko DB bulk inserts ya external API batch limits ke hisaab se chunk karta hai.
- 🛡️ **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Functional Filtering & Transformation
```go
package main

import (
	"fmt"
	"strings"
	gostream "go-stream"
)

type User struct {
	ID     int
	Name   string
	Active bool
}

func main() {
	users := []User{
		{ID: 1, Name: "Aman", Active: true},
		{ID: 2, Name: "Priya", Active: false},
		{ID: 3, Name: "Rohan", Active: true},
	}

	// Filter active users and extract their uppercase names
	activeNames := gostream.Map(
		gostream.FromSlice(users).Filter(func(u User) bool {
			return u.Active
		}),
		func(u User) string {
			return strings.ToUpper(u.Name)
		},
	).ToSlice()

	// Output: [AMAN ROHAN]
	fmt.Println("Active Users:", activeNames)
}
```

### 2. High-Performance Parallel Processing
```go
// Fetch 50 items concurrently using 8 background worker goroutines
enriched := gostream.ParallelMap(gostream.FromSlice(productIDs), 8, func(id int) ProductDetails {
    return fetchProductFromDatabaseOrAPI(id)
}).ToSlice()
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
