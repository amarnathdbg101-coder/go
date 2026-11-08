# go-exponentialbackoff

Mathematical exponential backoff and AWS-recommended jitter strategies for Go.

---

## Why Jitter?
Standard exponential backoff causes all retrying clients to wake up at the exact same time, creating massive spikes (**Thundering Herd Problem**). Jitter randomizes wait times across clients, smoothing out traffic spikes and maximizing recovery speed.

---

## Features
- **4 Jitter Algorithms**:
  - `FullJitter`: Maximum load reduction (`rand(0, ceiling)`).
  - `EqualJitter`: Guaranteed minimum delay + random spread.
  - `DecorrelatedJitter`: Independent history-based scaling.
  - `NoJitter`: Classic deterministic exponential curve.
- **Context-Aware**: `Sleep(ctx, attempt)` responds immediately to cancellations.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	goexponentialbackoff "go-exponentialbackoff"
)

func main() {
	backoff := goexponentialbackoff.New(goexponentialbackoff.Config{
		Base:      100 * time.Millisecond,
		Max:       5 * time.Second,
		Factor:    2.0,
		Algorithm: goexponentialbackoff.FullJitter,
	})

	for attempt := 0; attempt < 5; attempt++ {
		delay := backoff.Next(attempt)
		fmt.Printf("Attempt %d retry delay: %v\n", attempt, delay)

		// Or sleep with context awareness
		_ = backoff.Sleep(context.Background(), attempt)
	}
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
