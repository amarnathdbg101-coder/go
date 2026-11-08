# go-contextkit 🌐

Type-safe request context management, collision-free metadata accessors, detached background contexts, and HTTP tracing middleware for Go backends. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🛡️ **Zero Key Collisions**: Unexported private types context keys ke liye use hoti hain, avoiding library conflicts.
- 🚀 **Detached Context (`Detach`)**: Jab HTTP request khatam ho jati hai, to uska context cancel ho jata hai. `Detach(ctx)` background goroutines (email sending, async audit logs) ko bina cancel huye run karne deta hai jabki UserID aur RequestID preserve rehte hain!
- 🆔 **Request Tracing Middleware**: `X-Request-ID` aur Real Client IP (`X-Forwarded-For`) automatically inject aur echo karta hai.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Detached Async Background Task (Never Dies When HTTP Returns)
```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	ctxkit "go-contextkit"
)

func orderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = ctxkit.WithUserID(ctx, "USR-1092")

	// Fire background email without getting killed when HTTP client disconnects!
	bgCtx := ctxkit.Detach(ctx)
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Printf("Sent invoice email to user %s with trace %s\n",
			ctxkit.GetUserID(bgCtx), ctxkit.GetRequestID(bgCtx))
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"processing"}`))
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
