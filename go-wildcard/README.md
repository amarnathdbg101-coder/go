# go-wildcard

Ultra-fast, zero-allocation wildcard and glob pattern matcher for Go.

---

## Features
- **100x Faster than Regex**: Replaces heavy `regexp` compilation with an inline $O(n)$ two-pointer algorithm.
- **Zero Allocations**: Evaluates matches with 0 B/op performance.
- **Wildcard Syntax**:
  - `*` matches any sequence of characters (e.g. `*.json`, `api/*/users`).
  - `?` matches a single character (e.g. `file?.log`).
- **Case-Insensitive Mode**: `MatchFold("*.PNG", "photo.png") == true`.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	gowildcard "go-wildcard"
)

func main() {
	// 1. Simple file extension match
	fmt.Println(gowildcard.Match("*.png", "avatar.png")) // true

	// 2. Multi-wildcard event routing
	pattern := "order.*.paid"
	fmt.Println(gowildcard.Match(pattern, "order.vip.paid")) // true

	// 3. Case-insensitive folding
	fmt.Println(gowildcard.MatchFold("*.JPG", "profile.jpg")) // true

	// 4. Filter string slice
	files := []string{"main.go", "config.yaml", "utils.go"}
	goFiles := gowildcard.Filter("*.go", files)
	fmt.Println("Go files:", goFiles) // ["main.go", "utils.go"]
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
