# go-fsm

Thread-safe, generic Finite State Machine with guard conditions and lifecycle hooks for Go.

---

## Features
- **Type-Safe Generics**: Works with typed enums for States `S` and Events `E` without type casting.
- **Conditional Guard Predicates**: Attach `guard func() bool` closures to dynamically prevent unauthorized state changes.
- **Rich Lifecycle Callbacks**:
  - `OnEnter(state, hook)`: Executed when entering a state.
  - `OnExit(state, hook)`: Executed when exiting a state.
  - `OnTransition(hook)`: Global audit listener on every transition.
- **Introspection**: Query allowable transitions via `Can(event)` and `AvailableEvents()`.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	gofsm "go-fsm"
)

type OrderState string
type OrderEvent string

const (
	Pending   OrderState = "PENDING"
	Paid      OrderState = "PAID"
	Shipped   OrderState = "SHIPPED"
	Delivered OrderState = "DELIVERED"

	PayEvent     OrderEvent = "PAY"
	ShipEvent    OrderEvent = "SHIP"
	DeliverEvent OrderEvent = "DELIVER"
)

func main() {
	fsm := gofsm.New[OrderState, OrderEvent](Pending)

	// Define valid state transitions
	fsm.AddTransition(Pending, PayEvent, Paid, nil)
	fsm.AddTransition(Paid, ShipEvent, Shipped, nil)
	fsm.AddTransition(Shipped, DeliverEvent, Delivered, nil)

	// Register hooks
	fsm.OnEnter(Paid, func(from OrderState, event OrderEvent) {
		fmt.Println("Payment received! Generating invoice...")
	})

	// Trigger event
	if fsm.Can(PayEvent) {
		_ = fsm.Trigger(PayEvent)
	}

	fmt.Printf("Current Order State: %s\n", fsm.Current()) // PAID
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
