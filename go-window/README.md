# go-window 🪟

> **In-Memory Real-Time Stream Windowing Engine (Tumbling, Sliding, and Session Windows) with Watermarks for Go (Go 1.23+).**  
> Aggregate live event streams (fraud detection, login anomaly limits, real-time analytics) without running external streaming clusters like Apache Flink or Spark.

---

## 🚀 Key Features

1. **3 Production Window Topologies**:
   - **Tumbling**: Fixed, non-overlapping windows (e.g. $[10:00 - 10:05]$).
   - **Sliding**: Fixed-duration windows with regular slide steps (e.g. 5-min window sliding every 1-min).
   - **Session**: Dynamic windows bounded by periods of user inactivity.
2. **Watermarking & Late-Event Handling**:
   - Accommodates out-of-order events within `AllowedLateness`; safely drops corrupted historical anomalies.
3. **Pluggable Generic Reducer**:
   - Incrementally aggregates values (Sum, Max, Min, Average) using `WithReducer[T]`.
4. **Key-Partitioned Concurrency**:
   - Maintains independent window lifecycles per grouping key (e.g. `user_id`, `merchant_id`, `device_id`).
5. **Zero Dependencies**:
   - 100% pure standard library (`sync`, `time`).

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"time"
	"go-window"
)

func main() {
	// Configure a 1-minute tumbling window with a sum reducer
	cfg := window.TumblingConfig[int](1*time.Minute, func(res window.WindowResult[int]) {
		fmt.Printf("Window [%s - %s] for %s: Total Count=%d, Sum=%d\n",
			res.Start.Format("15:04:05"), res.End.Format("15:04:05"), res.Key, res.Count, res.Result)
	})

	engine, _ := window.New(cfg, window.WithReducer[int](func(a, b int) int {
		return a + b
	}))
	defer engine.Close()

	now := time.Now()
	// Add real-time transaction amounts
	_ = engine.Add(window.Event[int]{Key: "merchant_1", Timestamp: now, Value: 100})
	_ = engine.Add(window.Event[int]{Key: "merchant_1", Timestamp: now.Add(10 * time.Second), Value: 250})
}
```
