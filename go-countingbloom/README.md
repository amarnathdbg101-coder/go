# go-countingbloom 🌸

> **Counting Bloom Filter with Deletion Support for Go (Go 1.23+).**  
> Standard Bloom Filter me sabse bada limitation hota hai ki aap items ko delete nahi kar sakte (kyunki 1 bit ko 0 karne se dusre keys corrupt ho jaate hain). `go-countingbloom` 8-bit counters use karta hai jisse keys ko safely add (`Add`) aur remove (`Remove`) kiya ja sakta hai.

---

## 🚀 Features

- **Safe Deletion (`Remove`)**: Kisi item ke delete hone par counters decrement hote hain, aur baki keys affect nahi hoti hain.
- **Double-Hashing**: Kirsch-Mitzenmacher optimization (FNV-1a + SplitMix64) se $K$ hashes ultra-fast generate hote hain.
- **Saturation Protection**: Counter 255 par saturate ho jata hai (wrap around hokar 0 nahi banta), jisse false negatives ka risk 0 rehta hai.
- **Concurrent-Safe**: Multi-goroutine environment ke liye `NewConcurrent` RWMutex support.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-countingbloom"
)

func main() {
	// 10,000 items, 1% false positive rate
	filter, _ := countingbloom.New(10000, 0.01)

	// Items add karein
	filter.AddString("user:101")
	filter.AddString("user:102")

	fmt.Println("Contains user:101?", filter.ContainsString("user:101")) // true

	// Item delete karein! (Standard Bloom Filter me impossible)
	filter.RemoveString("user:101")

	fmt.Println("After Remove, Contains user:101?", filter.ContainsString("user:101")) // false
	fmt.Println("Contains user:102?", filter.ContainsString("user:102"))               // true
}
```
