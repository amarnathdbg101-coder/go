# go-shadow 👥

> **Asynchronous HTTP Traffic Shadowing, Mirroring & Response Diffing for Go (Go 1.23+).**  
> Test new microservice versions (v2) and database migrations against real production traffic without adding latency to primary callers or risking customer-facing bugs.

---

## 🚀 Key Features

1. **Zero Latency Penalty**:
   - Clones requests using non-blocking buffers; primary users receive their responses without waiting for the shadowed candidate service.
2. **Configurable Traffic Sampling**:
   - Sample fraction (e.g. `WithSampleRate(0.15)` for 15% of traffic) to control resource overhead on candidate clusters.
3. **Automated Response Comparison**:
   - Compares HTTP status codes, headers, and JSON body payloads, capturing precise diff reports upon discrepancy.
4. **Bounded Worker Queue**:
   - Protects host memory: drops shadow tasks under excessive load without affecting the primary API.
5. **Standard `net/http` Middleware**:
   - Plugs directly into `chi`, `net/http`, or any Go web router.
6. **Zero Dependencies**:
   - 100% pure Go standard library (`net/http`, `sync`, `time`, `bytes`, `io`).

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"net/http"
	"go-shadow"
)

func main() {
	// Configure shadow engine pointing to Candidate v2 service
	shadowEngine := shadow.New(
		shadow.WithTargetURL("http://candidate-v2.internal:8081"),
		shadow.WithSampleRate(0.20), // Mirror 20% of requests
		shadow.WithOnDiff(func(diff *shadow.DiffReport) {
			fmt.Printf("🚨 V2 DISCREPANCY on %s: Primary Status=%d, Candidate Status=%d\n",
				diff.Path, diff.PrimaryStatus, diff.CandidateStatus)
		}),
	)
	defer shadowEngine.Close()

	primaryMux := http.NewServeMux()
	primaryMux.HandleFunc("/api/checkout", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"order_completed","version":"v1"}`))
	})

	// Wrap with shadow middleware
	http.ListenAndServe(":8080", shadow.Middleware(shadowEngine)(primaryMux))
}
```
