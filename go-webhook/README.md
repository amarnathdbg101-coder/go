# go-webhook

Enterprise-grade outbound webhook dispatcher and inbound HMAC signature verifier for Go.

---

## Features
- **Standard HMAC-SHA256 Signing**: Compliant with Stripe, GitHub, and Shopify webhook header standards (`t=1700000000,v1=...`).
- **Anti-Replay Protection**: Built-in timestamp verification guards against replay and man-in-the-middle attacks.
- **Timing Attack Immune**: Constant-time `hmac.Equal` byte comparison.
- **Smart Retries**: Automatic exponential backoff retries for 5xx and 429 errors; immediately halts on permanent 4xx errors (e.g. 400, 401).
- **Concurrent Worker Pool**: Bounded goroutine queue for high-throughput asynchronous delivery without resource exhaustion.

---

## Quickstart

### 1. Inbound Webhook Verification (Consumer/Receiver)
```go
package main

import (
	"io"
	"net/http"
	"time"

	gowebhook "go-webhook"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	secret := []byte("whsec_your_customer_secret")
	body, _ := io.ReadAll(r.Body)
	sigHeader := r.Header.Get("X-Webhook-Signature")

	// Verify HMAC signature and ensure timestamp is not older than 5 minutes
	if err := gowebhook.Verify(secret, body, sigHeader, 5*time.Minute); err != nil {
		http.Error(w, "Invalid signature or expired webhook", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"received":true}`))
}
```

### 2. Outbound Webhook Dispatching (Publisher/Sender)
```go
package main

import (
	"context"
	"fmt"

	gowebhook "go-webhook"
)

func main() {
	dispatcher := gowebhook.NewDispatcher(gowebhook.DefaultConfig())
	defer dispatcher.Stop()

	secret := []byte("whsec_your_customer_secret")
	payload := []byte(`{"event":"order.paid","order_id":"ord_1001","amount":149950}`)

	// Synchronous delivery with 3 retries
	res, err := dispatcher.Dispatch(
		context.Background(),
		"https://customer-api.com/webhooks",
		"order.paid",
		"wh_evt_89234",
		payload,
		secret,
	)

	if err != nil {
		fmt.Printf("Webhook failed after %d attempts: %v\n", res.Attempts, err)
		return
	}

	fmt.Printf("Delivered successfully in %s (HTTP %d)\n", res.Duration, res.StatusCode)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
