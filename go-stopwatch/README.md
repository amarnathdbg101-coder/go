# go-stopwatch

High-resolution execution profiler and multi-lap stopwatch timer for Go.

---

## Features
- **Multi-Lap Stage Profiling**: Measure individual and cumulative split execution times across distinct stages.
- **Monotonic Precision**: Immune to system clock skew and NTP synchronization adjustments.
- **Visual Breakdown Summary**: Generates pretty formatted ASCII tables with percentage distribution.
- **JSON Telemetry Ready**: `ToMap()` exports structured metrics ready for structured logging.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"
	"time"

	gostopwatch "go-stopwatch"
)

func main() {
	sw := gostopwatch.Start("Order_Checkout_Pipeline")

	// Phase 1: Cart Validation
	time.Sleep(15 * time.Millisecond)
	sw.Lap("cart_validation")

	// Phase 2: Payment Gateway
	time.Sleep(85 * time.Millisecond)
	sw.Lap("payment_gateway")

	// Phase 3: PostgreSQL Database Insert
	time.Sleep(30 * time.Millisecond)
	sw.Lap("db_insert")

	sw.Stop()

	// Print visual summary
	fmt.Println(sw.Summary())
}
```

### Sample Output
```text
=== Stopwatch: Order_Checkout_Pipeline (Total: 130.5ms) ===
  [1] cart_validation      :     15.2ms ( 11.6%) [split: 15.2ms]
  [2] payment_gateway      :     85.1ms ( 65.2%) [split: 100.3ms]
  [3] db_insert            :     30.2ms ( 23.1%) [split: 130.5ms]
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
