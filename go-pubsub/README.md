# go-pubsub 📢

Generic type-safe in-memory pub-sub engine with channel-based fanout and slow-consumer non-blocking overflow mitigation. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🚀 **Non-Blocking Fanout**: Ek slow subscriber (jaise slow 2G mobile client) poore server ya doosre clients ko freeze nahi karta.
- ⚡ **Go Generics `[T any]`**: Strongly-typed payload transmission directly via `<-chan T`.
- 🚯 **Zero-Leak Unsubscribe**: Jab WebSocket client disconnect hota hai, to `sub.Unsubscribe()` channel ko gracefully close aur clean kar deta hai.
- 🎯 **Overflow Strategies**:
  - `DropOldest`: Purani unread message discard karke latest message save karta hai (ideal for real-time live GPS & stock tickers).
  - `DropNew`: Buffer full hone par new message drop karta hai.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. WebSocket Chat Room / Real-Time Order Tracking
```go
package main

import (
	"fmt"
	gopubsub "go-pubsub"
)

type LocationUpdate struct {
	DriverID string
	Lat      float64
	Lng      float64
}

func main() {
	// PubSub for real-time delivery tracking with DropOldest strategy
	ps := gopubsub.New[LocationUpdate](50, gopubsub.DropOldest)
	defer ps.Close()

	// Customer subscribes to order #102 updates
	sub := ps.Subscribe("order:102:tracking")
	defer sub.Unsubscribe()

	// Delivery agent publishes fresh GPS coordinates
	ps.Publish("order:102:tracking", LocationUpdate{
		DriverID: "DRV-99",
		Lat:      28.6139,
		Lng:      77.2090,
	})

	// Read update
	update := <-sub.Ch
	fmt.Printf("Driver %s is at (%f, %f)\n", update.DriverID, update.Lat, update.Lng)
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
