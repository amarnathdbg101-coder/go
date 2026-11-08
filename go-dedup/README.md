# go-dedup 🛡️

`go-dedup` ek ultra-fast, thread-safe, generic in-memory deduplication engine hai Go ke liye. Ye distributed webhooks (Stripe, Razorpay, GitHub, Shopify), event streams (Kafka, RabbitMQ, SQS), aur idempotent API endpoints ke duplicate events ko sub-microsecond time me filter out karta hai bina heavy database locks ya Redis roundtrips ke.

---

## ✨ Features

- **Generic Comparable Keys**: Kisi bhi type ke saath kaam karta hai (`string`, `int64`, UUIDs, custom comparable keys).
- **Atomic Check & Record**: `CheckAndRecord(id)` single atomic lock step me check aur record karta hai to eliminate race conditions.
- **Sliding-Window TTL**: Har event ka nanosecond timestamp maintain hota hai with automatic expiration.
- **Per-Key Custom TTL**: Standard window ke alawa dynamic TTL support (`CheckAndRecordWithTTL`).
- **Automatic Background Janitor**: Memory leak protection ke liye configurable background garbage collection.
- **Ultra-Low Memory Footprint**: Compact `map[K]int64` structure without pointer bloat.
- **Zero Third-Party Dependencies**: Pure Go Standard Library (`sync`, `time`).

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-dedup
```

### Webhook Deduplication Example
```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/shopsilo/go-dedup"
)

// In-memory deduplicator with 10-minute sliding window and 1-minute janitor cleanup
var webhookDeduper = dedup.New[string](10*time.Minute, 1*time.Minute)

func handlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
	eventID := r.Header.Get("X-Webhook-ID")
	if eventID == "" {
		http.Error(w, "missing event id", http.StatusBadRequest)
		return
	}

	// Atomically verify if this event was already processed
	if webhookDeduper.CheckAndRecord(eventID) {
		// Duplicate event detected! Return HTTP 200 OK immediately
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ignored_duplicate"}`))
		return
	}

	// First time seeing this event - proceed with business logic
	fmt.Printf("Processing new payment event: %s\n", eventID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"processed"}`))
}
```

### Manual Controls & Lifecycle
```go
d := dedup.New[string](5*time.Minute, 30*time.Second)
defer d.Close()

// Check without recording
if d.IsDuplicate("msg-99") {
    fmt.Println("Already seen!")
}

// Explicit purge & count
active := d.Count()
fmt.Printf("Active unique keys in window: %d\n", active)

// Explicit remove (e.g. if event failed processing and needs immediate retry)
d.Remove("msg-99")
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestDeduplicator_CheckAndRecord
--- PASS: TestDeduplicator_CheckAndRecord (0.12s)
=== RUN   TestDeduplicator_CustomTTL
--- PASS: TestDeduplicator_CustomTTL (0.07s)
=== RUN   TestDeduplicator_RemoveAndClear
--- PASS: TestDeduplicator_RemoveAndClear (0.00s)
=== RUN   TestDeduplicator_PurgeExpired
--- PASS: TestDeduplicator_PurgeExpired (0.08s)
=== RUN   TestDeduplicator_BackgroundJanitor
--- PASS: TestDeduplicator_BackgroundJanitor (0.10s)
=== RUN   TestDeduplicator_ConcurrentAccess
--- PASS: TestDeduplicator_ConcurrentAccess (0.00s)
PASS
ok  	go-dedup	0.865s
```

---

## 📄 License
MIT License.
