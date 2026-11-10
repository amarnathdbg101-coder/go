# go-cuckoofilter

`go-cuckoofilter` provides a high-performance Cuckoo Filter implementation in pure Go. Invented by Fan et al. (CMU, 2014), Cuckoo Filters are a superior modern alternative to Bloom Filters, supporting **true dynamic deletions**, lower false-positive rates, and higher memory utilization without counting-filter overhead.

## Key Features

- **True Dynamic Deletion**: Call `Delete(item)` safely without degrading the filter or increasing false positive probabilities.
- **16-bit High-Resolution Fingerprints**: Yields extremely low theoretical false positive rates ($< 0.01\%$).
- **4-Slot Bucket Associativity**: Achieves $> 95\%$ space load factor before saturation.
- **Cuckoo Displacement Relocation**: Resolves hash collisions dynamically via partial-key cuckoo kicking.
- **Thread-Safe**: RWMutex protected for concurrent high-throughput workloads.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-cuckoofilter"
)

func main() {
	cf, err := cuckoofilter.New(cuckoofilter.WithCapacity(100_000))
	if err != nil {
		panic(err)
	}

	// Insert items
	_ = cf.InsertString("user_session_101")
	_ = cf.InsertString("user_session_102")

	// Lookup
	fmt.Println("Contains 101:", cf.LookupString("user_session_101")) // true
	fmt.Println("Contains 999:", cf.LookupString("user_session_999")) // false

	// Dynamic Deletion (Impossible with standard Bloom Filters!)
	cf.DeleteString("user_session_101")
	fmt.Println("Contains 101 after delete:", cf.LookupString("user_session_101")) // false
	fmt.Printf("Current filter count: %d\n", cf.Count())                          // 1
}
```
