# go-chaos 💥

> **In-Process Chaos Engineering & Fault Injection Engine for Go (Go 1.23+).**  
> Test microservice resilience, circuit breakers, and timeout handlers by injecting latency, HTTP errors (500/502/503), payload corruption, and synthetic panics without external proxies.

---

## 🚀 Key Features

1. **4 Injected Fault Modalities**:
   - `LATENCY`: Injects delay and jitter to test client timeouts and circuit breakers.
   - `ERROR`: Injects custom HTTP status codes (500, 502, 503, 504) and JSON payloads.
   - `PANIC`: Injects controlled panics to test HTTP recovery middleware.
   - `CORRUPT`: Writes garbled response payloads to verify deserialization error handling.
2. **Granular Targeting**:
   - Path prefix (e.g. `/api/v1/payment/*`).
   - Header matching (e.g. `X-User-Tier: beta`).
   - Probability sampling (e.g. `0.20` = 20% of matching requests).
3. **Emergency Bypass & Master Killswitch**:
   - Instantly disable chaos with `engine.SetEnabled(false)`.
   - Bypass via `X-Bypass-Chaos: secret_token` headers for critical health checks.
4. **Standard `net/http` Middleware**:
   - Plugs directly into standard Go `http.Handler`, `chi`, `mux`, etc.
5. **Zero Dependencies**:
   - 100% Go standard library (`sync`, `time`, `net/http`, `crypto/rand`).

---

## 📦 Quickstart

```go
package main

import (
	"net/http"
	"time"

	"go-chaos"
)

func main() {
	engine := chaos.New()

	// Inject 500ms latency on 25% of payment API requests
	engine.AddRule(chaos.LatencyRule("/api/v1/payments", 500*time.Millisecond, 0.25))

	// Inject 503 Service Unavailable on 10% of checkout calls
	engine.AddRule(chaos.ErrorRule("/api/v1/checkout", http.StatusServiceUnavailable, 0.10))

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/payments", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"payment_processed"}`))
	})

	// Wrap with chaos middleware
	http.ListenAndServe(":8080", chaos.Middleware(engine)(mux))
}
```
