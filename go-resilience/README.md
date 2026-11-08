# go-resilience 🛡️

`go-resilience` ek unified, composable fault-tolerance aur resilience pipeline engine hai Go microservices ke liye. Ye Concurrency Bulkheading, Circuit Breaking, Exponential Backoff Retries with Full Jitter, Execution Timeouts, aur Graceful Fallbacks ko ek single fluent execution pipeline (`Policy[T]`) me unite karta hai.

---

## ✨ Features

- **All-In-One Resiliency**: Retries + Circuit Breaker + Bulkhead + Timeout + Fallback seamlessly composed.
- **Circuit Breaker Integration**: Closed, Open, aur Half-Open canary recovery states to protect downstream services from cascading outages.
- **Full Jitter Exponential Backoff**: Eliminates thundering herds on recovery.
- **Bulkhead Concurrency Guard**: Rejects excess traffic when downstream capacity is saturated (`ErrBulkheadFull`).
- **Graceful Fallback Degradation**: Cached data ya default responses return karein when dependencies are unreachable.
- **Type-Safe Generics**: Parameterized by return type `T`.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-resilience
```

### Complete Microservice Gateway Example
```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shopsilo/go-resilience"
)

func main() {
	// Build unified fault-tolerance policy
	paymentPolicy := resilience.NewPolicy[string]().
		WithTimeout(2 * time.Second).
		WithRetry(resilience.RetryConfig{
			MaxAttempts:     3,
			InitialInterval: 50 * time.Millisecond,
			Jitter:          true,
		}).
		WithCircuitBreaker(resilience.CircuitBreakerConfig{
			FailureThreshold: 3,
			ResetTimeout:     10 * time.Second,
		}).
		WithBulkhead(10). // max 10 concurrent requests
		WithFallback(func(ctx context.Context, err error) (string, error) {
			// Gracefully degrade to offline voucher or queued response
			return "QUEUED_OFFLINE", nil
		})

	// Execute call
	ctx := context.Background()
	result, err := paymentPolicy.Execute(ctx, func(ctx context.Context) (string, error) {
		// Call remote payment gateway
		return "PAYMENT_SUCCESS", nil
	})

	if err != nil {
		fmt.Printf("Execution failed: %v\n", err)
	} else {
		fmt.Printf("Payment Result: %s\n", result)
	}
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestPolicy_HappyPath
--- PASS: TestPolicy_HappyPath (0.00s)
=== RUN   TestPolicy_RetryTransientSuccess
--- PASS: TestPolicy_RetryTransientSuccess (0.03s)
=== RUN   TestPolicy_Timeout
--- PASS: TestPolicy_Timeout (0.03s)
=== RUN   TestPolicy_Bulkhead
--- PASS: TestPolicy_Bulkhead (0.05s)
=== RUN   TestPolicy_CircuitBreaker
--- PASS: TestPolicy_CircuitBreaker (0.00s)
=== RUN   TestPolicy_Fallback
--- PASS: TestPolicy_Fallback (0.01s)
PASS
ok  	go-resilience	0.606s
```

---

## 📄 License
MIT License.
