# go-rbtree

`go-rbtree` provides a thread-safe, generic Left-Leaning Red-Black (LLRB) self-balancing binary search tree in pure Go (invented by Robert Sedgewick, Princeton University). It maintains strict 2-3 tree isomorphism, guaranteeing $O(\log N)$ worst-case search, insertion, and floor/ceiling range queries.

## Key Features

- **Strict Balanced Height**: Left-leaning red link invariants guarantee tree height never exceeds $2 \log_2(N + 1)$, preventing degradation even under worst-case sorted insertions.
- **Type-Safe Generics**: Works over any ordered key type (`K cmp.Ordered`) and arbitrary value payload (`V any`).
- **Floor & Ceiling Queries**: Instant lookups for nearest bounding keys (`Floor` and `Ceiling`).
- **Sub-microsecond Range Scans**: In-order traversal (`AscendRange`) without heap allocations.
- **Thread-Safe**: RWMutex protected for concurrent read/write safety.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-rbtree"
)

func main() {
	tr := rbtree.New[int, string]()

	// Insert elements
	tr.Insert(100, "₹100 Coupon")
	tr.Insert(250, "₹250 Coupon")
	tr.Insert(500, "₹500 Coupon")

	// Point Lookup
	if val, ok := tr.Get(250); ok {
		fmt.Printf("Found: %s\n", val)
	}

	// Floor & Ceiling
	floorKey, floorVal, _ := tr.Floor(300)
	fmt.Printf("Nearest Coupon <= 300: Key %d -> %s\n", floorKey, floorVal) // 250

	// Range Scan [100..400]
	tr.AscendRange(100, 400, func(k int, v string) bool {
		fmt.Printf("  Key %d: %s\n", k, v)
		return true
	})
}
```
