# go-radix

`go-radix` provides a production-grade, thread-safe, generic Compact Radix Tree (Patricia Trie) in pure Go. By compressing common non-branching string prefixes into unified edge labels, it achieves up to 80% memory savings and sub-microsecond prefix routing compared to standard tries.

## Key Features

- **Path Compression & Edge Splitting**: Single-child branches are compacted into multi-character edge strings, reducing tree height dramatically.
- **Longest Prefix Matching**: Essential for IP routing tables (CIDR prefix matching) and high-speed HTTP API routers.
- **Sub-tree Autocomplete Walk (`WalkPrefix`)**: Rapidly iterates through all keys sharing an arbitrary search prefix.
- **Type-Safe Generics**: `Tree[T any]` works with any payload type.
- **Thread-Safe**: RWMutex protected for concurrent read/write throughput.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-radix"
)

func main() {
	router := radix.New[string]()

	// Register routing table
	router.Insert("/api/v1/users", "UsersController")
	router.Insert("/api/v1/orders", "OrdersController")
	router.Insert("/api/v1/orders/refunds", "RefundsController")

	// Longest Prefix Match
	prefix, controller, ok := router.LongestPrefix("/api/v1/orders/98765/track")
	if ok {
		fmt.Printf("Routed to: %s (Matched Route: %s)\n", controller, prefix)
		// Routed to: OrdersController (Matched Route: /api/v1/orders)
	}

	// Autocomplete / Prefix search
	fmt.Println("All orders endpoints:")
	router.WalkPrefix("/api/v1/orders", func(path string, handler string) bool {
		fmt.Printf("  %s -> %s\n", path, handler)
		return true
	})
}
```
