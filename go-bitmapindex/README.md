# go-bitmapindex

`go-bitmapindex` provides a high-performance Columnar Bit-Sliced Index (BSI) engine in pure Go. Used in analytical databases (OLAP), data warehouses, and fast analytics engines, Bit-Sliced Indexing executes range filters and mathematical aggregations (`SUM`, `AVG`, `COUNT`) using hardware bitwise instructions (`POPCNT`) without scanning raw table rows.

## Key Features

- **Columnar Bit-Sliced Layout**: Stores numeric columns decomposed into orthogonal bit slices ($S_0, S_1, \dots, S_{B-1}$).
- **Hardware-Accelerated Aggregation**: Computes column `SUM` using $B$ hardware population counts (`bits.OnesCount64`) instead of looping over millions of individual rows.
- **Fast Range Filtering**: Evaluates range filters (`val BETWEEN min AND max`) using boolean bit operations.
- **Null Safety**: Dedicated existence bitset tracks populated rows without memory waste.
- **Thread-Safe**: RWMutex protected for concurrent analytical queries and writes.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-bitmapindex"
)

func main() {
	// 16-bit integer index (supports values up to 65,535)
	idx, err := bitmapindex.New(16)
	if err != nil {
		panic(err)
	}

	// Index order transaction amounts (rowID -> amount)
	_ = idx.Set(101, 1500)
	_ = idx.Set(102, 2500)
	_ = idx.Set(103, 800)
	_ = idx.Set(104, 3200)

	// Range filter: orders between 1000 and 3000
	matchingRows, _ := idx.FilterRange(1000, 3000)
	fmt.Printf("Orders in range ₹1000..₹3000: %v\n", matchingRows) // [101, 102]

	// Instant SUM across matching orders via bitwise POPCNT
	filteredSum := idx.Sum(matchingRows)
	fmt.Printf("Total Revenue for matching orders: ₹%d\n", filteredSum) // ₹4000
}
```
