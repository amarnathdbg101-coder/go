# go-keyrotator 🔑

Thread-safe API Key Pooling, Round-Robin Rotation, and Automatic 429 Cooldown Failover for Go.

## 🚀 Key Features
- **Zero External Dependencies**: 100% Pure Go standard library (`sync`, `time`, `strings`).
- **HTTP 429 Resilience**: Marks keys in cooldown and automatically routes subsequent requests to alternative healthy keys.
- **Circuit-Breaker Error Threshold**: Automatically quarantines problematic keys after consecutive failures.
- **Earliest Recovery Fallback**: If all keys are saturated, returns the key that will recover earliest so callers can decide whether to wait or fail.

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"time"

	"github.com/shopsilo/go-keyrotator"
)

func main() {
	// Initialize from CSV environment variable or slice
	rotator := keyrotator.FromCSV("AIzaSyA...,AIzaSyB...,AIzaSyC...")

	// 1. Fetch next healthy key in round-robin sequence
	key, err := rotator.GetNextKey()
	if err != nil {
		panic(err)
	}

	// 2. Perform API call...
	success := true
	if success {
		rotator.MarkSuccess(key)
	} else {
		// If received HTTP 429 Too Many Requests:
		rotator.MarkRateLimited(key, 5*time.Minute)
	}

	fmt.Printf("Key used: %s\n", key)
}
```
