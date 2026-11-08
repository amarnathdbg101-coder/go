# go-lfu ⚡

`go-lfu` ek ultra-fast, thread-safe $O(1)$ Least Frequently Used (LFU) cache library hai Go ke liye. LRU ke opposite, ye access frequency track karta hai taaki long-term popular items (e-commerce hot products, search keywords, auth tokens) temporary request spikes ke dauraan evict hone se protected rahein.

---

## ✨ Features

- **True $O(1)$ Time Complexity**: Doubly-linked frequency bucket lists provide constant-time `Get` and `Set`.
- **Intelligent Tie-Breaking**: Same frequency hone par oldest least-recently-used item ko evict karta hai.
- **Per-Item TTL Expiration**: Support for dynamic TTL expiration (`SetWithTTL`).
- **Synchronous Batch Purging**: `PurgeExpired()` removes stale keys without blocking live requests.
- **Type-Safe Generics**: Works seamlessly with any comparable key `K` and value `V`.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-lfu
```

### Complete Example
```go
package main

import (
	"fmt"
	"time"

	"github.com/shopsilo/go-lfu"
)

func main() {
	// Cache with max capacity 3
	cache := lfu.New[string, string](3)

	cache.Set("hot_item", "iPhone 15")
	cache.Set("item_b", "Laptop Bag")
	cache.Set("item_c", "USB Cable")

	// Access "hot_item" multiple times to boost frequency
	cache.Get("hot_item")
	cache.Get("hot_item")
	cache.Get("hot_item")

	// Access "item_b" once
	cache.Get("item_b")

	// Adding 4th item evicts "item_c" (frequency 1)
	cache.Set("item_d", "Mechanical Keyboard")

	if _, ok := cache.Get("item_c"); !ok {
		fmt.Println("item_c was evicted because it was least frequently used!")
	}

	// Set with TTL
	cache.SetWithTTL("otp_code", "9921", 5*time.Minute)
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestLFU_Eviction
--- PASS: TestLFU_Eviction (0.00s)
=== RUN   TestLFU_TieBreaking
--- PASS: TestLFU_TieBreaking (0.00s)
=== RUN   TestLFU_TTL
--- PASS: TestLFU_TTL (0.06s)
=== RUN   TestLFU_PurgeExpired
--- PASS: TestLFU_PurgeExpired (0.05s)
=== RUN   TestLFU_DeleteAndClear
--- PASS: TestLFU_DeleteAndClear (0.00s)
=== RUN   TestLFU_Concurrency
--- PASS: TestLFU_Concurrency (0.01s)
PASS
ok  	go-lfu	0.647s
```

---

## 📄 License
MIT License.
