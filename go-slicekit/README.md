# go-slicekit 🔪

`go-slicekit` ek high-performance, type-safe generic slice manipulation toolkit hai Go 1.22+ ke liye. Ye Go standard library me missing functional & slice utility functions provide karta hai jaise `Chunk`, `Flatten`, `Unique`, `GroupBy`, `Partition`, `Compact`, `Shuffle`, `Map`, `Filter`, aur `Reduce`.

---

## ✨ Features

- **Chunking & Flattening**: Batch API calls and DB queries with `Chunk(items, 100)` and `Flatten(batches)`.
- **Order-Preserving Deduplication**: `Unique(slice)` aur custom key extractor `UniqueBy(slice, keyFn)`.
- **Classification & Splitting**: `GroupBy(users, keyFn)` and `Partition(items, predicate)`.
- **Functional Pipeline**: `Map`, `Filter`, `Reduce`, `Find`, `FindIndex`, `ContainsBy`.
- **Reversals & Compacting**: `Reverse`, `ReverseInPlace`, `Compact` (strips zero-values).
- **Crypto-Secure Shuffle**: Fisher-Yates randomization via `crypto/rand`.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-slicekit
```

### Complete Example
```go
package main

import (
	"fmt"

	"github.com/shopsilo/go-slicekit"
)

type Product struct {
	ID       int
	Category string
	Price    float64
}

func main() {
	// 1. Chunk large ID list into batches of 3
	ids := []int{101, 102, 103, 104, 105, 106, 107}
	batches := slicekit.Chunk(ids, 3)
	fmt.Printf("Batches: %v\n", batches) // [[101 102 103] [104 105 106] [107]]

	// 2. Deduplicate while preserving order
	tags := []string{"go", "backend", "go", "api", "backend"}
	uniqueTags := slicekit.Unique(tags)
	fmt.Printf("Unique Tags: %v\n", uniqueTags) // ["go", "backend", "api"]

	// 3. Group by category
	catalog := []Product{
		{ID: 1, Category: "Electronics", Price: 1200},
		{ID: 2, Category: "Books", Price: 450},
		{ID: 3, Category: "Electronics", Price: 800},
	}
	grouped := slicekit.GroupBy(catalog, func(p Product) string {
		return p.Category
	})
	fmt.Printf("Electronics count: %d\n", len(grouped["Electronics"])) // 2

	// 4. Map & Filter
	prices := slicekit.Map(catalog, func(p Product) float64 {
		return p.Price
	})
	fmt.Printf("Prices: %v\n", prices) // [1200, 450, 800]
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestChunk
--- PASS: TestChunk (0.00s)
=== RUN   TestFlatten
--- PASS: TestFlatten (0.00s)
=== RUN   TestUniqueAndUniqueBy
--- PASS: TestUniqueAndUniqueBy (0.00s)
=== RUN   TestGroupBy
--- PASS: TestGroupBy (0.00s)
=== RUN   TestPartition
--- PASS: TestPartition (0.00s)
=== RUN   TestReverseAndCompact
--- PASS: TestReverseAndCompact (0.00s)
=== RUN   TestMapFilterReduce
--- PASS: TestMapFilterReduce (0.00s)
=== RUN   TestFindAndContainsBy
--- PASS: TestFindAndContainsBy (0.00s)
=== RUN   TestShuffle
--- PASS: TestShuffle (0.00s)
PASS
ok  	go-slicekit	0.484s
```

---

## 📄 License
MIT License.
