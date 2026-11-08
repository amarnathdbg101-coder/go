# go-bloom 🌸

Sub-50ns, zero-allocation probabilistic Bloom filter for cache penetration defense, duplicate checking, and set membership queries. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- ⚡ **Blazing Speed**: ~42 ns per lookup, **0 B/op** and **0 heap allocations**!
- 🛡️ **Cache Penetration Shield**: Non-existent product IDs ya usernames ko database hit karne se pehle hi sub-microsecond me reject karta hai.
- 🧮 **Kirsch-Mitzenmacher Double Hashing**: FNV-1a aur SplitMix64 ke zariye $O(1)$ me $k$ independent hash probes generate karta hai without memory overhead.
- 💾 **Serialization**: `Bytes()` aur `FromBytes()` ke zariye filter ko disk par persist ya Redis me cache kiya ja sakta hai.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Defending PostgreSQL & Redis Against Cache Penetration
```go
package main

import (
	"fmt"
	"log"
	gobloom "go-bloom"
)

func main() {
	// Sized for 1,000,000 product IDs with 1% false positive rate (~1.2MB RAM)
	bf, err := gobloom.New(1000000, 0.01)
	if err != nil {
		log.Fatal(err)
	}

	// Load valid product IDs on startup
	bf.AddString("PROD-1001")
	bf.AddString("PROD-1002")

	// Query Check
	incomingID := "PROD-9999" // Malicious / non-existent ID
	if !bf.ContainsString(incomingID) {
		// Guaranteed 100% that it does not exist in DB!
		// Return 404 immediately without touching Redis or PostgreSQL!
		fmt.Printf("Item %s definitely does NOT exist. Blocked DB query!\n", incomingID)
		return
	}

	// Only query DB if Bloom filter says it might exist
	fmt.Printf("Item %s might exist. Proceeding to Redis/DB.\n", incomingID)
}
```

---

## Running Tests & Benchmarks
```bash
go test -v -bench=. -benchmem ./...
```
