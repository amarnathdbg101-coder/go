# go-syncchan ⚡

> **Leak-Proof Concurrency Combinators for Go Channels (Go 1.23+ Generics).**  
> Go me channels ke sath kaam karte waqt goroutine leak sabse common bug hota hai (jaise agar reader cancel ho gaya par writer channel par block reh gaya). `go-syncchan` production-grade, zero-allocation generic combinators deta hai jo context cancellation ke sath completely leak-proof hain.

---

## 🚀 Features

- **FanIn**: Multiple channels ko ek consolidated output stream me merge karta hai.
- **Tee**: Ek single channel stream ko bina race condition ke do independent channels me split karta hai.
- **FanOut**: Ek producer se $N$ parallel consumer channels tak data broadcast karta hai.
- **Take & Drop**: Stream me se pehle $N$ elements lena ya pehle $N$ elements drop karna.
- **OrDone Guard**: Context cancel hote hi channel reader goroutine ko gracefully terminate karta hai.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"context"
	"fmt"
	"go-syncchan"
)

func main() {
	ctx := context.Background()

	// 1. Do alag worker channels
	ch1 := make(chan string, 2)
	ch2 := make(chan string, 2)

	ch1 <- "Kafka Event"
	ch2 <- "RabbitMQ Event"
	close(ch1)
	close(ch2)

	// 2. FanIn se dono ko ek stream me merge karein
	merged := syncchan.FanIn(ctx, ch1, ch2)

	// 3. Process merged stream
	for msg := range merged {
		fmt.Println("Received:", msg)
	}
}
```
