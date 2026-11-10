# go-tdigest

`go-tdigest` provides a production-grade implementation of Ted Dunning's t-Digest algorithm in pure Go. It calculates accurate streaming quantiles and percentiles (P50, P90, P95, P99, P99.9) across unbounded numeric telemetry data (HTTP request latencies, network metrics, financial tickers) using bounded constant memory ($< 4\text{ KB}$).

## Key Features

- **Extreme Tail Accuracy**: The t-Digest scale function enforces smaller centroid clusters at the tails ($q \to 0$ and $q \to 1$), delivering sub-0.1% error for P99 and P99.9 latency SLOs.
- **Mergeable Online Summaries**: Multiple t-Digests can be merged (`Merge`) across distributed worker nodes with zero loss of accuracy.
- **O(1) Memory Budget**: Clustered into $\le 2\delta$ centroids ($\approx 100$ to $200$ centroids total) regardless of whether you process $1,000$ or $100,000,000$ events.
- **Thread-Safe**: RWMutex protected for concurrent telemetry ingestion and quantile reads.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-tdigest"
)

func main() {
	// Initialize t-Digest with compression delta 100
	td, err := tdigest.New(tdigest.WithCompression(100))
	if err != nil {
		panic(err)
	}

	// Stream microservice API latencies in milliseconds
	for i := 1; i <= 10000; i++ {
		td.Add(float64(i)) // 1ms .. 10,000ms
	}

	// Calculate latency percentiles
	p50, _ := td.Quantile(0.50) // Median
	p90, _ := td.Quantile(0.90) // P90
	p99, _ := td.Quantile(0.99) // P99

	fmt.Printf("Latency SLA Metrics:\n")
	fmt.Printf("  P50: %.2f ms\n", p50)
	fmt.Printf("  P90: %.2f ms\n", p90)
	fmt.Printf("  P99: %.2f ms\n", p99)
}
```
