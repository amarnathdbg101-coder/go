# go-metrics 📈

Zero-dependency, lock-free in-memory Prometheus-compatible metrics engine with atomic counters, gauges, histograms (p50/p95/p99), Go runtime samplers, and HTTP middleware. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🚀 **Lock-Free Atomics**: Counters aur Gauges me 0-lock atomic operations (`CompareAndSwapUint64`) use hoti hain for maximum CPU throughput.
- 📊 **Prometheus Ready**: Official Prometheus exposition text format (v0.0.4) me output deta hai for Grafana / Prometheus scrapers.
- ⏱️ **Histograms & Percentiles**: Configurable buckets ke saath latency distribution aur approximate `p50`, `p90`, `p99` calculations.
- 🩺 **Runtime Telemetry**: Goroutines count, allocated heap memory, aur GC pauses automatically monitor karta hai.
- 🌐 **HTTP Auto-Instrumentation**: Drop-in middleware for Chi / net/http to record request counts aur latency.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. HTTP Server Instrumentation
```go
package main

import (
	"net/http"
	gometrics "go-metrics"
)

func main() {
	reg := gometrics.NewRegistry()
	sampler := gometrics.RegisterRuntimeMetrics(reg)

	mux := http.NewServeMux()

	// Business API
	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order_id": 101}`))
	})

	// Prometheus Scrape Endpoint
	mux.HandleFunc("/metrics", gometrics.Handler(reg, sampler))

	// Wrap server with latency & count tracking middleware
	handler := gometrics.HTTPMiddleware(reg)(mux)

	http.ListenAndServe(":8080", handler)
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
