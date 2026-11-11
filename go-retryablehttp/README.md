# go-retryablehttp 🔄

> [!NOTE]
> **Consolidation Notice**: Yeh package ab full resilience features (Body Rewinding via `GetBody`, `Retry-After` header parsing, `OnRetry` observability callbacks, aur Circuit Breaker) ke sath **[`go-httpclient`](../go-httpclient)** me consolidate ho chuka hai. Naye projects ke liye `go-httpclient` ka use recommend kiya jata hai. Existing code backward-compatible rahega.

> **Resilient, Zero-Dependency HTTP Client Wrapper for Go (Go 1.23+).**  
> Go ke standard `net/http` client me agar POST request retry karni pade, toh pehli attempt ke baad body drain ho jati hai aur retry fail ho jati hai. `go-retryablehttp` request body ko auto-rewind karta hai, `Retry-After` headers ko parse karta hai, aur full-jitter backoff ke sath safe microservice communication deta hai.

---

## 🚀 Features

- **Body Rewind Guarantee**: POST, PUT, aur PATCH payloads ko safe copy karke har retry attempt me rewind karta hai (`req.GetBody`).
- **`Retry-After` Header Parsing**: 429 Too Many Requests ya 503 me server dwara bheje gaye seconds ya HTTP-date ko respect karta hai.
- **Full Jitter Backoff**: Downstream services par thundering herd attack rokne ke liye randomized exponential backoff.
- **Leak-Proof Sockets**: Har retry se pehle pichle response body ko drain aur close karta hai taaki connection pool starve na ho.
- **Context-Aware**: Parent context cancel ya timeout hote hi turant retry execution stop ho jati hai.

---

## 📦 Quickstart

```go
package main

import (
	"context"
	"fmt"
	"go-retryablehttp"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	// 1. Resilient client configure karein
	client := retryablehttp.New(
		retryablehttp.WithMaxRetries(3),
		retryablehttp.WithWaitLimits(100*time.Millisecond, 2*time.Second),
		retryablehttp.WithOnRetry(func(attempt int, req *http.Request, resp *http.Response, err error) {
			fmt.Printf("⚠️ Attempt %d failed. Retrying...\n", attempt)
		}),
	)

	// 2. Body-rewindable request banayein
	payload := `{"order_id": "ORD-1234", "amount": 499}`
	req, _ := retryablehttp.NewRequest(context.Background(), http.MethodPost, "https://api.example.com/checkout", strings.NewReader(payload))

	// 3. Request execute karein (automatic retry on 5xx, 429, network drop)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d, Response: %s\n", resp.StatusCode, string(body))
}
```
