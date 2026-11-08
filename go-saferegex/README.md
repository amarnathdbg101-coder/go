# go-saferegex 🛡️

> **ReDoS-Protected Regular Expression Matcher & LRU Cache for Go (Go 1.23+).**  
> Malicious user input ya complex regex patterns backend servers ko 100% CPU lockup me daal sakte hain (Regular Expression Denial of Service - ReDoS). `go-saferegex` execution ko strict deadline timeout ke sath isolate karta hai aur compiled patterns ko LRU cache me maintain karta hai.

---

## 🚀 Features

- **ReDoS Protection**: Regex evaluation ko background goroutine aur hard timeout boundary ke sath wrap karta hai; agar pattern freeze hota hai toh `ErrTimeout` return hota hai.
- **LRU Pattern Cache**: Bar-bar compile hone wale regex patterns ko memory me cache karta hai taaki compilation latency $0$ ho jaye.
- **Context-Aware**: Parent HTTP/gRPC request context cancel hone par matching process turant terminate ho jata hai.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-saferegex"
	"time"
)

func main() {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	email := "arjun@example.com"

	// 50ms ke strict timeout ke sath execute karein
	matched, err := saferegex.MatchString(pattern, email, 50*time.Millisecond)
	if err != nil {
		fmt.Println("Validation error or timeout:", err)
		return
	}

	fmt.Printf("Is Valid Email: %v\n", matched)
}
```
