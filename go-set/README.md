# go-set 🎯

`go-set` ek high-performance, type-safe generic `Set[T comparable]` data structure library hai Go ke liye. Ye manual `map[T]struct{}` boilerplate ko replace karta hai with rich set algebra (Union, Intersect, Difference, Symmetric Difference), functional helpers (`Filter`, `Each`), JSON serialization, aur optional thread-safety modes.

---

## ✨ Features

- **Type-Safe Generics**: Kisi bhi Go comparable type ke saath natively compatible (`string`, `int`, custom IDs).
- **Dual Concurrency Modes**:
  - `New()`: Maximum performance for local/single-routine operations (0 mutex overhead).
  - `NewThreadSafe()`: `RWMutex`-protected instance for concurrent goroutines.
- **Set Algebra Suite**: Union, Intersection, Difference ($A \setminus B$), Symmetric Difference ($A \Delta B$), Subset & Superset.
- **Functional Helpers**: `Filter(predicate)`, `Each(fn)`, `ToSlice()`, `Clone()`.
- **JSON Serialization**: Direct JSON API response marshaling and unmarshaling.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-set
```

### Complete Example
```go
package main

import (
	"fmt"

	"github.com/shopsilo/go-set"
)

func main() {
	// Active user permissions
	adminRoles := set.New("read", "write", "delete", "export")
	editorRoles := set.New("read", "write", "upload")

	// 1. Intersection: Common permissions
	common := adminRoles.Intersect(editorRoles)
	fmt.Printf("Common permissions: %v\n", common.ToSlice()) // ["read", "write"]

	// 2. Difference: Admin-only permissions
	adminOnly := adminRoles.Difference(editorRoles)
	fmt.Printf("Admin only: %v\n", adminOnly.ToSlice()) // ["delete", "export"]

	// 3. Union: Combined permission set
	all := adminRoles.Union(editorRoles)
	fmt.Printf("Total unique permissions: %d\n", all.Size()) // 5

	// 4. Membership check
	if adminRoles.Contains("delete") {
		fmt.Println("Admin can delete records!")
	}
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestSet_BasicOps
--- PASS: TestSet_BasicOps (0.00s)
=== RUN   TestSet_Algebra
--- PASS: TestSet_Algebra (0.00s)
=== RUN   TestSet_FilterAndEach
--- PASS: TestSet_FilterAndEach (0.00s)
=== RUN   TestSet_JSON
--- PASS: TestSet_JSON (0.00s)
=== RUN   TestSet_ThreadSafety
--- PASS: TestSet_ThreadSafety (0.00s)
PASS
ok  	go-set	0.571s
```

---

## 📄 License
MIT License.
