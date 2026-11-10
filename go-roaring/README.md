# go-roaring

`go-roaring` provides a high-performance, compressed 32-bit Roaring Bitmap data structure in pure Go. Roaring Bitmaps are widely used in modern search engines (Lucene, Elasticsearch) and analytical databases (ClickHouse, Pinot, Spark) to index sets of integers with massive memory compression and sub-microsecond set operations.

## Key Features

- **Adaptive Dual-Container Architecture**:
  - **Array Container**: Used for sparse data (< 4,096 elements). Stores 16-bit sorted integers in contiguous memory for minimal overhead.
  - **Bitmap Container**: Automatically promoted when cardinality reaches $\ge$ 4,096 elements. Represents 65,536 integers in exactly 1,024 64-bit words ($8\text{ KB}$ fixed buffer).
- **Hardware-Accelerated Bitwise Operations**: Uses Go standard library `math/bits.OnesCount64` (POPCNT instruction) and bitwise manipulation.
- **Fast Set Theory**: Instant `And` (intersection), `Or` (union), and `AndNot` (difference) across millions of integer IDs.
- **Thread-Safe**: RWMutex protected for concurrent read/write operations.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-roaring"
)

func main() {
	bmA := roaring.New()
	bmB := roaring.New()

	// Add user IDs matching criteria A
	bmA.AddMany(10, 20, 30, 100000)

	// Add user IDs matching criteria B
	bmB.AddMany(20, 30, 40, 100000)

	// Intersect: Users matching both A AND B
	intersection := bmA.And(bmB)
	fmt.Println("Intersection count:", intersection.Cardinality()) // 3: [20, 30, 100000]

	// Union: Users matching A OR B
	union := bmA.Or(bmB)
	fmt.Println("Union count:", union.Cardinality()) // 5: [10, 20, 30, 40, 100000]
}
```
