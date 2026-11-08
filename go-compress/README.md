# go-compress 🗜️

Recycled `sync.Pool` HTTP gzip compression middleware with size threshold buffering and standalone streaming utilities. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- ♻️ **Recycled Buffer Pool**: `sync.Pool` ke zariye `gzip.Writer` allocations reuse hoti hain, eliminating GC pauses on high-traffic servers.
- 📏 **Threshold Buffering (MinSize)**: Chhoti responses (< 1KB) ko compress nahi karta, preventing gzip header bloat on tiny JSON status payloads.
- 📶 **Mobile Data Saver**: Flutter mobile apps aur web SPAs ke liye 60-80% bandwidth reduce karta hai.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Attaching to HTTP Router
```go
package main

import (
	"net/http"
	gocompress "go-compress"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		// Large JSON response (> 1KB automatically gets gzipped)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(largeProductCatalogJSON))
	})

	// Wrap server with gzip middleware
	http.ListenAndServe(":8080", gocompress.Middleware()(mux))
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
