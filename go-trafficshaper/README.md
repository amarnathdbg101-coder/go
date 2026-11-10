# go-trafficshaper

`go-trafficshaper` provides a production-grade Leaky Bucket and Token Bucket traffic shaper with priority admission queues in pure Go. It smoothly shapes erratic request bursts into an evenly paced stream, preventing downstream rate-limit violations (such as HTTP 429 Too Many Requests from third-party APIs or message broker queue overflows).

## Key Features

- **Dual Algorithms**:
  - **Token Bucket**: Allows controlled short-term bursts up to a burst capacity while sustaining average throughput.
  - **Leaky Bucket**: Strictly enforces a constant emission interval with zero burst tolerance.
- **Priority Queue Scheduling**: High-priority tasks (e.g. payment transactions or critical webhooks) jump ahead of low-priority background jobs.
- **Bounded Wait Buffer**: Protects server memory by rejecting excess incoming traffic with `ErrQueueFull`.
- **Context & Deadline Cancellation**: Callers can pass standard `context.Context` to abort gracefully if wait times exceed SLA deadlines.
- **Real-Time Telemetry**: Tracks admitted, queued, dropped, and current queue depth.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"time"

	"go-trafficshaper"
)

func main() {
	// Paces requests to external payment gateway at max 50 req/sec with burst of 10
	shaper, err := trafficshaper.New(
		trafficshaper.WithRate(50.0),
		trafficshaper.WithBurst(10),
		trafficshaper.WithQueueCapacity(100),
	)
	if err != nil {
		panic(err)
	}
	defer shaper.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// High-priority payment event
	if err := shaper.AcquireWithPriority(ctx, 10); err != nil {
		fmt.Printf("Payment request dropped or timed out: %v\n", err)
		return
	}

	fmt.Println("Payment request safely dispatched within upstream rate limits!")
}
```
