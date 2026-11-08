# go-bitmap

Ultra-compact, high-performance bit array / bitset with hardware POPCNT and set algebra for Go.

---

## Features
- **98.4% Memory Savings**: Replaces `map[int]bool` (~16 bytes per entry) with 1 bit per entry (1 million flags = ~122 KB RAM).
- **Hardware-Accelerated POPCNT**: Counts set bits in nanoseconds using CPU assembly instructions via `math/bits.OnesCount64`.
- **Set Algebra**: Fast bitwise operations (`And` / intersection, `Or` / union, `AndNot` / difference).
- **Dynamic Growth**: Automatically expands buffer when setting bits beyond initial capacity.
- **Binary Serialization**: Clean export and import via `ToBytes()` and `FromBytes()`.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	gobitmap "go-bitmap"
)

func main() {
	// Track 1,000,000 active user IDs using only ~122 KB RAM
	activeUsers := gobitmap.New(1000000)

	// Mark user 42 and 9999 as active
	_ = activeUsers.Set(42)
	_ = activeUsers.Set(9999)

	// Check status in 1 CPU cycle
	fmt.Println("Is user 42 active?", activeUsers.Test(42))     // true
	fmt.Println("Is user 100 active?", activeUsers.Test(100))   // false

	// Count total active users using hardware POPCNT
	fmt.Printf("Total active users: %d\n", activeUsers.Count()) // 2

	// Fast set intersection (e.g. Users active in both Group A and Group B)
	groupA := gobitmap.New(100)
	groupB := gobitmap.New(100)
	_ = groupA.Set(5)
	_ = groupA.Set(10)
	_ = groupB.Set(10)
	_ = groupB.Set(20)

	inBoth := groupA.And(groupB)
	fmt.Println("Users in both groups:", inBoth.Count()) // 1 (User 10)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
