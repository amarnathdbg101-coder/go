# go-ratelimit 🛡️

> **High-Performance, Production-Grade Rate Limiting Library for Go.**  
> Zero-dependency Token Bucket & Sliding Window algorithms, thread-safe, 0ms overhead, with pluggable In-Memory & Redis cluster stores.

---

## ✨ Features

- ⚡ **Ultra-Fast Performance**: In-memory evaluation in **< 50 nanoseconds**.
- 🪣 **Token Bucket & Sliding Window**: Accurate capacity limits with burst support.
- 🛡️ **Zero Memory Leaks**: Automated background ticker cleanup for inactive keys.
- 🌐 **Pluggable Storage (`Store`)**:
  - `MemoryStore`: Zero-setup, thread-safe in-memory store for instant local testing and single-node apps.
  - `RedisStore`: Distributed rate limiting for microservice clusters using atomic Lua scripts.
- 🔑 **Flexible Key Extraction (`KeyFunc`)**: Rate limit by Client IP, User ID, API Key, Request Path, or composite keys.
- 📡 **Standard HTTP Headers**: RFC 6585 compliant headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After`).
- 🔌 **Universal HTTP Middleware**: Compatible with Chi, Gin, Fiber, Echo, Gorilla, or standard `net/http`.

---

## 📦 Installation

```bash
go get github.com/shopsilo/go-ratelimit
```

---

## ⚡ 3-Minute Quickstart

### 1. Simple IP-Based Rate Limiting (100 req / minute)

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/shopsilo/go-ratelimit"
)

func main() {
	// Create rate limiter: 100 requests per minute by Client IP
	limiter, err := ratelimit.NewIPLimiter(100, 1*time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	defer limiter.Close()

	// Wrap your standard http.Handler / Router
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success": true}`))
	})

	log.Println("Server running at :8080...")
	http.ListenAndServe(":8080", limiter.Middleware(mux))
}
```

---

## 🔐 Advanced Usage & Key Strategies

### Rate Limit by API Key or Header

```go
limiter, _ := ratelimit.New(
    ratelimit.NewMemoryStore(),
    ratelimit.Rule{Capacity: 50, Window: 1 * time.Minute},
    ratelimit.WithKeyFunc(ratelimit.KeyByHeader("X-API-Key")),
)
```

### Rate Limit Login Endpoint (Prevent Brute-Force Password Attacks)

```go
// Limit /login specifically by Path + IP: max 5 login attempts per minute
loginLimiter, _ := ratelimit.New(
    ratelimit.NewMemoryStore(),
    ratelimit.Rule{Capacity: 5, Window: 1 * time.Minute},
    ratelimit.WithKeyFunc(ratelimit.KeyByPathAndIP),
)

r.With(loginLimiter.Middleware).Post("/login", handleLogin)
```

---

## 🧪 Testing

Run comprehensive unit tests:

```bash
cd C:\Library\go-ratelimit
go test -v ./...
```
