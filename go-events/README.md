# go-events 📡

High-performance, in-memory event bus and pub-sub engine with wildcard pattern matching, async worker pools, isolated panic recovery, and Dead-Letter Queue (DLQ). Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🎯 **Topic Pattern Wildcards**: Dot-separated namespace matching (`order.*`, `user.**`, `payment.>` or exact match).
- ⚡ **Dual Dispatch**: Synchronous `Publish()` aur background bounded worker-pool `PublishAsync()`.
- 🛡️ **Panic Isolation**: Agar kisi ek event listener me panic aaye, to wo catch ho jati hai aur baki listeners bina ruke execute hote hain.
- 📬 **Dead-Letter Queue (DLQ)**: Agar koi event bina subscriber ke ho ya handler error de, to wo DLQ callback me capture hota hai.
- 🪢 **Middleware Pipeline**: Correlation ID injection, tracing, aur latency measurement ke liye pluggable middlewares.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Synchronous & Asynchronous Event Pub-Sub
```go
package main

import (
	"context"
	"fmt"
	"log"
	goevents "go-events"
)

type OrderCreatedPayload struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

func main() {
	// Initialize Bus with 4 worker threads
	bus := goevents.New(goevents.Config{
		WorkerCount: 4,
		QueueSize:   1024,
	})
	defer bus.Close()

	// Wildcard subscription: listens to any event starting with 'order.'
	sub, _ := bus.Subscribe("order.*", func(ctx context.Context, e goevents.Event) error {
		payload := e.Payload.(OrderCreatedPayload)
		fmt.Printf("Received %s -> Order: %s, Amount: ₹%.2f\n", e.Topic, payload.OrderID, payload.Amount)
		return nil
	})
	defer sub.Unsubscribe()

	// Dispatch event
	event := goevents.NewEvent("order.created", OrderCreatedPayload{
		OrderID: "ORD-88123",
		Amount:  4999.00,
	})

	_ = bus.Publish(context.Background(), event)
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
