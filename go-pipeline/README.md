# go-pipeline

Generic, composable, onion-model middleware and workflow execution pipeline for Go.

---

## Features
- **Type-Safe Generics**: Works seamlessly with any user-defined struct (`*Order`, `*UserContext`) without reflection.
- **Onion-Model Execution**: Upstream stages can wrap around downstream executions to perform timing, audit logging, or transaction management.
- **Immediate Short-Circuiting**: Early returns or errors halt execution immediately.
- **Context-Aware**: Respects `ctx.Done()` at each stage boundary.
- **Panic Isolation**: Protects your application from crashing if a third-party step panics.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	gopipeline "go-pipeline"
)

type CheckoutContext struct {
	UserID string
	CartTotal int
	IsPaid bool
}

func main() {
	pipe := gopipeline.New[CheckoutContext]()

	// 1. Timing & Logging Middleware (Onion Model)
	pipe.Use(func(ctx context.Context, c *CheckoutContext, next gopipeline.NextFunc) error {
		start := time.Now()
		fmt.Println("[PIPELINE] Starting checkout for:", c.UserID)
		err := next(ctx)
		fmt.Printf("[PIPELINE] Finished in %s, error: %v\n", time.Since(start), err)
		return err
	})

	// 2. Sequential Validation Step
	pipe.AddStep(func(ctx context.Context, c *CheckoutContext) error {
		if c.CartTotal <= 0 {
			return fmt.Errorf("cart total must be greater than zero")
		}
		return nil
	})

	// 3. Payment Processing Step
	pipe.AddStep(func(ctx context.Context, c *CheckoutContext) error {
		fmt.Println("Charging customer card...")
		c.IsPaid = true
		return nil
	})

	order := &CheckoutContext{UserID: "usr_998", CartTotal: 1500}
	if err := pipe.Execute(context.Background(), order); err != nil {
		fmt.Printf("Checkout aborted: %v\n", err)
		return
	}

	fmt.Printf("Order completed successfully: %+v\n", order)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
