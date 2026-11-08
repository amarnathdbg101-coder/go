# go-deepclone 🧬

`go-deepclone` ek fast, reflection-safe deep cloning engine hai Go ke liye. Ye Go structures, slices, maps, nested pointers, aur cyclic graphs ka completely independent deep copy banata hai bina kisi accidental mutation side-effects ya data races ke.

---

## ✨ Features

- **True Memory Independence**: Nested pointers, slices, aur maps ka naya memory allocate karta hai (modifying the clone never mutates the original).
- **Circular Reference Protection**: Cyclic graphs (e.g. `node.Next = node`) ko detect karta hai bina infinite recursion ya stack overflow ke.
- **Type-Safe Generics**: `Clone[T](src T) T` returns the exact compile-time type without interface casting.
- **Nil Fidelity**: Nil pointers, nil slices, aur nil maps ko accurately nil preserve karta hai.
- **Faster than JSON Roundtrip**: Bypasses slow `json.Marshal` + `Unmarshal` strings and preserves exact types.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-deepclone
```

### Complete Example
```go
package main

import (
	"fmt"

	"github.com/shopsilo/go-deepclone"
)

type Config struct {
	AppName  string
	Features []string
	Database *DBConfig
}

type DBConfig struct {
	Host string
	Port int
}

func main() {
	original := Config{
		AppName:  "OrderService",
		Features: []string{"checkout", "coupons"},
		Database: &DBConfig{Host: "localhost", Port: 5432},
	}

	// Create deep independent clone
	clone := deepclone.Clone(original)

	// Mutate clone
	clone.Features[0] = "MUTATED"
	clone.Database.Host = "remote-db"

	// Original remains pristine!
	fmt.Printf("Original Features: %v | DB: %s\n", original.Features, original.Database.Host)
	// Output: Original Features: [checkout coupons] | DB: localhost

	fmt.Printf("Clone Features: %v | DB: %s\n", clone.Features, clone.Database.Host)
	// Output: Clone Features: [MUTATED coupons] | DB: remote-db
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestClone_StructWithPointersAndCollections
--- PASS: TestClone_StructWithPointersAndCollections (0.00s)
=== RUN   TestClone_CircularReference
--- PASS: TestClone_CircularReference (0.00s)
=== RUN   TestClone_NilPointersAndSlices
--- PASS: TestClone_NilPointersAndSlices (0.00s)
PASS
ok  	go-deepclone	0.487s
```

---

## 📄 License
MIT License.
