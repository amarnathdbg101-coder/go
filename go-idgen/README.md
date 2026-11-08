# go-idgen 🆔

Ultra-fast (~4.8M IDs/sec), zero-allocation distributed 64-bit Snowflake ID generator with Base62 encoding and clock drift protection. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🚀 **Blazing Speed**: ~240 ns per ID, **0 B/op** and **0 heap allocations** per generation.
- 🐘 **PostgreSQL `BIGINT` Optimized**: Generates 63-bit positive integers that fit directly into `BIGINT PRIMARY KEY` without negative numbers or fragmentation.
- 🕒 **Time-Ordered**: Database B-tree index locality ensures fast inserts and natural chronological sorting.
- 🌐 **Distributed Ready**: Supports up to 1,024 independent machine/container nodes.
- 🛡️ **Clock Drift Protection**: NTP backwards adjustments par duplicate IDs prevent karta hai.
- 🔤 **Base62 Encoding**: Short alphanumeric strings (`7xK9pL2`) for user-facing order numbers and short URLs.
- 📦 **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Generating Distributed 64-bit IDs
```go
package main

import (
	"fmt"
	"log"
	goidgen "go-idgen"
)

func main() {
	// Initialize Node with NodeID = 1 (Range: 0 - 1023)
	node, err := goidgen.NewNode(1)
	if err != nil {
		log.Fatal(err)
	}

	// Generate 64-bit integer ID
	id, _ := node.Generate()
	fmt.Printf("Snowflake ID: %d\n", id)

	// Generate compact Base62 string for customer URLs/slips
	base62Str, _ := node.GenerateBase62()
	fmt.Printf("Base62 Order Reference: %s\n", base62Str)

	// Extract timestamp from ID without database query
	createdAt, nodeID, _ := node.Deconstruct(id)
	fmt.Printf("Created At: %s on Node: %d\n", createdAt, nodeID)
}
```

---

## Running Tests & Benchmarks
```bash
go test -v -bench=. -benchmem ./...
```
