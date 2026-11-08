# go-slidingstats 📊

> **Rolling Window Streaming Statistics & Percentiles Engine for Go (Go 1.23+).**  
> Microservices me API latency percentiles (P50, P95, P99), dynamic load monitoring, aur moving statistics (Min, Max, Mean, StdDev) track karne ke liye ring-buffer histogram engine. Bina millions of floats ko RAM me store kiye sliding window statistics calculate karta hai.

---

## 🚀 Features

- **Summary Statistics**: `Count`, `Min`, `Max`, `Mean`, `StdDev` instant computation.
- **High-Precision Percentiles**: `P50` (Median), `P90`, `P95`, aur `P99` moving percentiles.
- **Ring Buffer Time Windows**: Rolling time window (e.g. 1 minute with 10 sub-buckets) automatically old data discard karta hai.
- **Bounded RAM Footprint**: Per-bucket sample caps ke through memory usage strictly bounded rehti hai.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-slidingstats"
	"time"
)

func main() {
	// 1-minute window with 10 sub-buckets (6s each)
	stats, _ := slidingstats.New(1*time.Minute, 10)

	// API Latencies (milliseconds) observe karein
	stats.Observe(12.5)
	stats.Observe(18.2)
	stats.Observe(24.0)
	stats.Observe(120.5) // spike
	stats.Observe(15.1)

	// Current rolling stats snapshot lein
	snap := stats.Snapshot()
	fmt.Printf("Requests: %d\n", snap.Count)
	fmt.Printf("Min: %.1fms, Max: %.1fms, Mean: %.1fms\n", snap.Min, snap.Max, snap.Mean)
	fmt.Printf("P50: %.1fms, P95: %.1fms, P99: %.1fms\n", snap.P50, snap.P95, snap.P99)
}
```
