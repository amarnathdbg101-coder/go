# go-btree

`go-btree` provides a high-performance, thread-safe, generic B-Tree in pure Go. It delivers $O(\log N)$ point lookups, insertions, deletions, and high-speed bidirectional range scans while maximizing CPU L1/L2 cache locality via wide, cache-friendly node branching.

## Key Features

- **Generic Types**: Type-safe over any ordered key type (`K cmp.Ordered`) and arbitrary value payload (`V any`).
- **CPU Cache Locality**: High branching factor (default degree $t = 32$, up to 63 keys per node packed in contiguous arrays) eliminates pointer-chasing CPU cache misses common in binary search trees (AVL / Red-Black).
- **Single-Pass Preemptive Splitting**: Top-down node splitting eliminates recursion bubbling on insertion.
- **Fast Range Scans**: Sub-microsecond range iteration via `AscendRange(start, end, fn)`.
- **Thread-Safe**: RWMutex protected for concurrent reads and writes.
- **Zero Dependencies**: Pure Go standard library (`cmp`, `sort`, `sync`).

## Quick Start

```go
package main

import (
	"fmt"
	"go-btree"
)

func main() {
	// Create B-Tree with degree 32
	tr, err := btree.New[int, string](btree.WithDegree(32))
	if err != nil {
		panic(err)
	}

	// Insert key-values
	tr.Set(1001, "Order #1001: ₹1,499")
	tr.Set(1005, "Order #1005: ₹3,200")
	tr.Set(1002, "Order #1002: ₹850")

	// Point Lookup
	if val, ok := tr.Get(1002); ok {
		fmt.Printf("Lookup 1002: %s\n", val)
	}

	// Range Scan [1001, 1004]
	fmt.Println("Range Scan [1001..1004]:")
	tr.AscendRange(1001, 1004, func(k int, v string) bool {
		fmt.Printf("Key: %d -> %s\n", k, v)
		return true
	})
}
```
