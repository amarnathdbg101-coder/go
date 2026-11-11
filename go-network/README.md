# go-network 🌐

> **Unified, High-Performance Networking, Traffic & Resiliency Suite for Go.**  
> Combines Rate Limiting, Resilient HTTP Client, Circuit Breaker, CORS, Healthchecks, and Graceful Shutdown into a single modular Go package.

---

## 📦 Sub-packages

- `github.com/shopsilo/go-network/ratelimit`: Token Bucket & Sliding Window Rate Limiter
- `github.com/shopsilo/go-network/httpclient`: Resilient net/http client with full-jitter retries
- `github.com/shopsilo/go-network/circuitbreaker`: Universal function & HTTP circuit breaker
- `github.com/shopsilo/go-network/ipkit`: Real client IP extractor & anti-spoofing
- `github.com/shopsilo/go-network/cors`: Zero-allocation CORS middleware

---

## ⚡ 3-Minute Quickstart

```go
package main

import (
	"net/http"
	"time"

	"github.com/shopsilo/go-network/ratelimit"
)

func main() {
	// Rate limit: 100 req / minute by Client IP
	limiter, _ := ratelimit.NewIPLimiter(100, 1*time.Minute)
	defer limiter.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success": true}`))
	})

	http.ListenAndServe(":8080", limiter.Middleware(mux))
}
```
