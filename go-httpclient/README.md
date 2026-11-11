# go-httpclient 🌐

Resilient, production-grade HTTP client with automated exponential backoff retries, full jitter, circuit breaker, and zero-leak connection pooling. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🔄 **Exponential Backoff & Full Jitter**: Network drops or 429/502/503 errors par automatically retry karta hai bina thundering herd create kiye.
- ⚡ **Circuit Breaker Built-in**: Downstream external service fail ho to cascading failure rokte hue fast-fail karta hai.
- 🚰 **Zero-Leak Guarantee**: Automatic response body draining (`io.Discard`) aur closing connection pool exhaust hone se bachata hai.
- 📦 **Fluent JSON API**: Struct ko directly encode aur response ko target pointer me decode karta hai.
- 🛡️ **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Simple GET Request with Automatic JSON Decoding
```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"
	gohttp "github.com/amarnathdbg101-coder/go/go-httpclient"
)

type WeatherResponse struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temp"`
}

func main() {
	client := gohttp.New(
		gohttp.WithTimeout(5 * time.Second),
		gohttp.WithHeader("User-Agent", "ShopSilo-Backend/1.0"),
	)

	var weather WeatherResponse
	ctx := context.Background()

	_, err := client.Get(ctx, "https://api.example.com/weather?city=Delhi", &weather)
	if err != nil {
		log.Fatalf("API request failed: %v", err)
	}

	fmt.Printf("City: %s, Temp: %.1f°C\n", weather.City, weather.Temperature)
}
```

### 2. POST with Circuit Breaker and Custom Retries
```go
retryCfg := gohttp.DefaultRetryConfig
retryCfg.MaxRetries = 3
retryCfg.InitialInterval = 200 * time.Millisecond

client := gohttp.New(
    gohttp.WithTimeout(10 * time.Second),
    gohttp.WithRetry(retryCfg),
    gohttp.WithCircuitBreaker(5, 30*time.Second), // 5 failures -> Trips breaker for 30s
)

orderPayload := map[string]any{
    "order_id": "ORD-9912",
    "amount":   2499.00,
}

var paymentResponse map[string]any
_, err := client.Post(ctx, "https://api.razorpay.com/v1/orders", orderPayload, &paymentResponse)
if err != nil {
    // If circuit is open, err will be gohttp.ErrCircuitOpen
    log.Println("Payment call failed:", err)
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
