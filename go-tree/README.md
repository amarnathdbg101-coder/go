# go-tree

Generic Prefix Trie & Hierarchical Category Tree engine for Go.

---

## Features
- **Prefix Trie**: High-speed prefix lookup (`StartsWith`) and autocompletion for search bars and URL routing.
- **Hierarchical Category Tree**: Multi-root tree with breadcrumbs (`Ancestors`), sub-tree extraction (`Descendants`), and cycle protection.
- **Generic Payloads**: Attach any typed payload `T` to nodes without reflection.
- **Thread Safe**: Protected with concurrent `sync.RWMutex`.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

### 1. Autocomplete Prefix Trie
```go
package main

import (
	"fmt"

	gotree "go-tree"
)

func main() {
	trie := gotree.NewTrie[string]()
	trie.Insert("apple", "Apple Fruit")
	trie.Insert("application", "Software App")
	trie.Insert("approach", "Method")

	// Autocomplete matching "app"
	matches := trie.StartsWith("app")
	fmt.Println("Matches:", matches) // ["Apple Fruit", "Software App", "Method"]
}
```

### 2. Category Hierarchy & Breadcrumbs
```go
package main

import (
	"fmt"

	gotree "go-tree"
)

func main() {
	h := gotree.NewHierarchy[string]()

	_ = h.Add("c_electronics", "", "Electronics")
	_ = h.Add("c_phones", "c_electronics", "Mobile Phones")
	_ = h.Add("c_smartphones", "c_phones", "Smartphones")

	// Get Breadcrumb Trail
	crumbs, _ := h.Ancestors("c_smartphones")
	for _, c := range crumbs {
		fmt.Printf("%s > ", c.Value)
	}
	fmt.Println("Smartphones")
	// Output: Electronics > Mobile Phones > Smartphones
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
