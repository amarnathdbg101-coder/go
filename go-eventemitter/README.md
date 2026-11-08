# go-eventemitter 📢

> **Thread-Safe Event Emitter with Wildcards & Leak Alerts for Go (Go 1.23+).**  
> Microservices, plugin systems, aur UI frameworks ke liye Node.js-style EventEmitter pattern. `On`, `Once`, `Emit`, `EmitAsync`, Wildcard events (`order.*`), aur goroutine memory leak warning threshold support.

---

## 🚀 Features

- **`On` & `Once` with Unsubscribe**: Har subscription ek clean `unsub()` function return karti hai.
- **Sync & Async Modes**: `Emit` synchronous dispatch karta hai jabki `EmitAsync` non-blocking goroutines me parallel triggers karta hai.
- **Wildcard Subscriptions**: `user.*` subscribe karke `user.login`, `user.logout` sabhi events capture karein.
- **Panic Isolation**: Agar koi ek listener code me panic karta hai, toh baki listeners affect nahi hote aur server crash nahi hota.
- **Listener Leak Warning**: Agar kisi event par 30+ listeners accumulate ho rahe hon, toh automated warning trigger hoti hai.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-eventemitter"
)

func main() {
	emitter := eventemitter.New()

	// 1. Wildcard subscription
	unsub, _ := emitter.On("order.*", func(event string, payload any) {
		fmt.Printf("Event [%s] received with data: %v\n", event, payload)
	})

	// 2. Events emit karein
	emitter.Emit("order.placed", map[string]any{"id": "ORD-101", "total": 999})
	emitter.Emit("order.delivered", map[string]any{"id": "ORD-101"})

	// 3. Unsubscribe
	unsub()
}
```
