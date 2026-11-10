# go-hyperloglog

`go-hyperloglog` provides a production-grade HyperLogLog++ (HLL++) cardinality estimation algorithm in pure Go. It estimates the number of distinct elements (e.g. daily active users, unique IP visitors, distinct search terms) across billions of streaming events with sub-1% relative standard error while using only 16 KB of memory.

## Key Features

- **64-bit Hash Resolution**: Avoids 32-bit hash collision saturation, scaling past $10^{18}$ unique items.
- **LinearCounting Small-Range Bias Correction**: Accurately counts small cardinalities ($< 2.5 \cdot m$) where standard raw harmonic estimates skew.
- **Precision Scalability**: Configurable precision $p \in [4, 18]$ ($m = 2^p$ registers). Default $p=14$ provides $\approx 0.81\%$ standard error using 16 KB RAM.
- **Distributed Merge (Union)**: Easily merge sketches computed independently on worker nodes or shards in $O(m)$ time with zero accuracy loss.
- **Thread-Safe**: RWMutex protected for concurrent real-time metric ingestion.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-hyperloglog"
)

func main() {
	// Create sketch with precision 14 (16,384 registers, ~0.81% error)
	hll, err := hyperloglog.New(hyperloglog.WithPrecision(14))
	if err != nil {
		panic(err)
	}

	// Stream 1 million items (with duplicates)
	for i := 0; i < 1_000_000; i++ {
		hll.AddString(fmt.Sprintf("user_id_%d", i%50_000)) // 50,000 unique users
	}

	fmt.Printf("Estimated unique users: %d (Actual: 50,000)\n", hll.Estimate())

	// Distributed Merge
	nodeA, _ := hyperloglog.New()
	nodeB, _ := hyperloglog.New()

	nodeA.AddString("user_1")
	nodeB.AddString("user_2")

	_ = nodeA.Merge(nodeB)
	fmt.Printf("Cluster-wide unique users: %d\n", nodeA.Estimate()) // 2
}
```
