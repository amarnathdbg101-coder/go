# go-orderedmap 🗺️

> **High-Performance, Zero-Dependency Generic Ordered Map for Go (Go 1.23+).**  
> Go ke native map me iteration order random hota hai. `go-orderedmap` $O(1)$ fast lookup provide karta hai aur strict insertion order maintain karta hai, saath hi first-class deterministic JSON serialization support karta hai.

---

## 🚀 Features

- **$O(1)$ Performance**: Hash table + Doubly linked list ka use karke Get, Set aur Delete teeno $O(1)$ time complexity me execute hote hain.
- **Go 1.23+ Generics**: Type-safe `Map[K comparable, V any]` bina kisi `interface{}` ya reflection overhead ke.
- **Deterministic JSON**: `json.Marshal` aur `json.Unmarshal` keys ka exact sequence preserve karte hain (configuration files aur API contracts ke liye perfect).
- **Concurrency Support**: Single-threaded speed ke liye `New()` ya multi-goroutine safe ke liye `NewConcurrent()`.
- **In-Place Sorting**: Custom sorting function (`Sort`) ke sath entries ko reorder kar sakte hain.

---

## 📦 Quickstart

```go
package main

import (
	"encoding/json"
	"fmt"
	"go-orderedmap"
)

func main() {
	om := orderedmap.New[string, any]()

	// 1. Insertion order preserve hoga
	om.Set("user_id", "USR-9901")
	om.Set("name", "Arjun Sharma")
	om.Set("role", "Admin")

	// 2. Strict order me iterate karein
	om.ForEach(func(key string, val any) bool {
		fmt.Printf("%s => %v\n", key, val)
		return true
	})

	// 3. Deterministic JSON output
	data, _ := json.Marshal(om)
	fmt.Println(string(data))
	// Output: {"user_id":"USR-9901","name":"Arjun Sharma","role":"Admin"}
}
```
