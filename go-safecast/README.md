# go-safecast

Zero-allocation, bounds-checked, overflow-safe numeric casting for Go using generics.

---

## Why go-safecast?
Go's built-in type conversions fail silently on out-of-bounds numbers:
```go
// Native Go:
val := int8(300)     // Result: 44 ❌ (Silent data corruption!)
u := uint(int(-1))   // Result: 18446744073709551615 ❌ (Massive unsigned wraparound!)
```

`go-safecast` prevents these bugs with typed error returns:
```go
val, err := gosafecast.ToInt8(300) // err: gosafecast.ErrOverflow ✅
u, err := gosafecast.ToUint(-1)    // err: gosafecast.ErrUnderflow ✅
```

---

## Features
- **Generic Constraints**: Supports all standard numeric types (`int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`).
- **Zero Allocations**: Pure inline logic with 0 B/op performance.
- **NaN / Inf Protection**: Safely catches IEEE-754 `NaN` and `±Inf` without crashing.
- **Must Helper**: Includes `Must(v, err)` for configurations or guaranteed bounds.

---

## Quickstart

```go
package main

import (
	"fmt"

	gosafecast "go-safecast"
)

func main() {
	// 1. Safe downcasting from 64-bit int to 8-bit int
	var bigID int64 = 100
	id8, err := gosafecast.ToInt8(bigID)
	if err != nil {
		fmt.Println("Cast error:", err)
		return
	}
	fmt.Printf("Safe int8: %d\n", id8) // 100

	// 2. Catching overflow
	var overflowID int64 = 500
	_, err = gosafecast.ToInt8(overflowID)
	if err != nil {
		fmt.Println("Caught overflow successfully:", err) // gosafecast: numeric overflow
	}

	// 3. Catching negative unsigned conversion
	negative := -10
	_, err = gosafecast.ToUint(negative)
	if err != nil {
		fmt.Println("Caught negative underflow:", err) // gosafecast: numeric underflow
	}
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
