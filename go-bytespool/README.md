# go-bytespool 🌊

> **Tiered Zero-Allocation Byte Buffer Pool with Anti-Memory Leak Caps for Go (Go 1.23+).**  
> Go me standard `sync.Pool` ka sabse bada drawback "Buffer Bloat" hota hai — jab koi ek heavy request buffer ko 10MB expand kar deti hai, toh future me choti requests ko bhi wahi 10MB buffer milta hai, jisse memory leak jaisa RAM spike dikhta hai. `go-bytespool` 4 discrete size tiers (512B, 4KB, 64KB, 1MB) provide karta hai aur oversized buffers ko automatically discard karta hai.

---

## 🚀 Features

- **Tiered Architecture**: Capacity hint ke according appropriate pool se buffer fetch hota hai (Small, Medium, Large, Huge).
- **Anti-Bloat Protection**: Agar koi buffer maximum limit (>2MB) se bada ho jata hai, toh use pool me recycle karne ke bajaye discard kar diya jata hai taaki GC reclaim kar sake.
- **Zero-Allocation Formatting**: `WriteString`, `WriteInt`, `WriteFloat` direct buffer slice me append karte hain, bina kisi string ya heap allocation ke.
- **Scoped Helper**: `WithBuffer(size, func(b))` automatic `Get` aur deferred `Put` handle karta hai.

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-bytespool"
)

func main() {
	// Scoped buffer use karein (auto-returned to pool)
	bytespool.WithBuffer(1024, func(buf *bytespool.Buffer) {
		buf.WriteString("Order #")
		buf.WriteInt(98721)
		buf.WriteString(" Total: ₹")
		buf.WriteFloat(1499.50, 2)

		fmt.Println(buf.String())
		// Output: Order #98721 Total: ₹1499.50
	})
}
```
