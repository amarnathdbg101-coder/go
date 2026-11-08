# go-ipkit 🌐

Secure Real Client IP extractor with trusted proxy backward traversal, spoofing protection, and CIDR subnet evaluation. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🛡️ **Anti-IP-Spoofing**: Attacker dwara bheje gaye fake `X-Forwarded-For` headers ko reject karta hai agar direct connection kisi untrusted IP se ho.
- 🔄 **Trusted Proxy Traversal**: `X-Forwarded-For` chain ko backwards traverse karke true end-user IP nikalta hai.
- ☁️ **CDN Support**: Cloudflare (`CF-Connecting-IP`) aur Akamai (`True-Client-IP`) direct support.
- 🔍 **Subnet & CIDR Matching**: `IsPrivate(ip)` aur `InCIDR(ip, "10.0.0.0/8")` utilities.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Extracting Real Client IP in HTTP Middleware
```go
package main

import (
	"fmt"
	"net/http"
	goipkit "go-ipkit"
)

func main() {
	// Trust private VPC load balancers (AWS ALB / K8s ingress)
	extractor := goipkit.New(goipkit.WithTrustPrivate(true))

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		clientIP := extractor.GetRealIP(r)
		fmt.Printf("Real Client IP: %s (IsPrivate: %v)\n", clientIP, goipkit.IsPrivate(clientIP))
	})

	http.ListenAndServe(":8080", nil)
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
