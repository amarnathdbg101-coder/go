# go-heavyhitters

`go-heavyhitters` provides an in-memory implementation of the Space-Saving Algorithm (Metwally et al., ACM TODS 2005) for streaming Top-K Frequent Items ("Heavy Hitters") in pure Go. It identifies dominant elements across massive unbounded streams (trending topics, popular products, top IP traffic) using small, strictly bounded memory.

## Key Features

- **Space-Saving Algorithm**: Tracks Top-K items in streaming data with bounded memory $O(M)$ where $M = 1/\epsilon$.
- **Guaranteed Error Bounds**: Guarantees that any item with true frequency $> N/M$ is captured, with maximum overestimation error explicitly reported per item.
- **O(1) Streaming Increments**: Sub-microsecond execution time per item.
- **Thread-Safe**: RWMutex protected for concurrent high-speed stream processing.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-heavyhitters"
)

func main() {
	// Track heavy hitters using 100 slots
	tracker, _ := heavyhitters.New(heavyhitters.WithCapacity(100))

	// Stream occurrences
	tracker.AddWithCount("product_iphone_16", 12000)
	tracker.AddWithCount("product_macbook_pro", 8500)
	tracker.AddWithCount("product_airpods", 4000)

	// Stream noise
	for i := 0; i < 50000; i++ {
		tracker.Add(fmt.Sprintf("tail_product_%d", i))
	}

	// Query Top 3 trending products
	top3 := tracker.TopK(3)
	for i, item := range top3 {
		fmt.Printf("#%d %s: ~%d events (Max Error: %d)\n", 
			i+1, item.Key, item.Count, item.Error)
	}
}
```
