# go-timingwheel 🎡

`go-timingwheel` ek high-performance, $O(1)$ Hashed Timing Wheel timer scheduler library hai Go ke liye. Ye standard Go runtime `time.AfterFunc` ke heap contention ko eliminate karta hai aur hundreds of thousands of concurrent network timeouts (e.g. WebSockets, RPC connection timeouts, delayed queues) ko constant time me handle karta hai.

---

## ✨ Features

- **True $O(1)$ Operations**: Start aur Cancel operations constant-time doubly-linked list manipulation use karte hain.
- **Multi-Round Expiration**: Wheel length se lambi durations ko `rounds` tracking ke through accurately schedule karta hai.
- **Decoupled Execution**: Task execution separate goroutines me hoti hai taaki tick cycle lock-free aur uninterrupted rahe.
- **High Concurrency**: Millions of concurrent timers without GC heap pressure.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-timingwheel
```

### Complete Example
```go
package main

import (
	"fmt"
	"time"

	"github.com/shopsilo/go-timingwheel"
)

func main() {
	// Tick resolution: 100ms, Wheel size: 60 slots (6 seconds per revolution)
	tw := timingwheel.New(100*time.Millisecond, 60)
	tw.Start()
	defer tw.Stop()

	// 1. Schedule delayed task
	tw.AfterFunc(500*time.Millisecond, func() {
		fmt.Println("Heartbeat ping sent!")
	})

	// 2. Schedule and cancel connection timeout
	timer, _ := tw.AfterFunc(3*time.Second, func() {
		fmt.Println("Connection timed out!")
	})

	// Data received early -> cancel timeout in O(1)
	if timer.Stop() {
		fmt.Println("Timeout successfully cancelled!")
	}

	time.Sleep(1 * time.Second)
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestTimingWheel_BasicExecution
--- PASS: TestTimingWheel_BasicExecution (0.06s)
=== RUN   TestTimingWheel_Cancel
--- PASS: TestTimingWheel_Cancel (0.14s)
=== RUN   TestTimingWheel_MultiRound
--- PASS: TestTimingWheel_MultiRound (0.12s)
=== RUN   TestTimingWheel_Concurrency
--- PASS: TestTimingWheel_Concurrency (0.10s)
PASS
ok  	go-timingwheel	0.901s
```

---

## 📄 License
MIT License.
