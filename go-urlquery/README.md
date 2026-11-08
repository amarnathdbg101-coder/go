# go-urlquery 🔗

`go-urlquery` ek lightweight, bidirectional URL query parameter encoder aur decoder library hai Go structs ke liye. Ye HTTP handlers me `r.URL.Query()` ko strongly-typed filter structs me parse karne ke boilerplate ko completely eliminate karta hai.

---

## ✨ Features

- **Bidirectional Conversion**: Unmarshal `url.Values` to struct (`Decode`) aur marshal struct to query string (`Encode`).
- **Declarative Struct Tags**: `query:"page,default=1"`, `query:"tags,comma"`, `query:"ids"`.
- **Rich Type Conversions**: `string`, `int/8/16/32/64`, `float32/64`, `bool`, `time.Duration`, and slices.
- **Comma & Multi-Value Slices**: Supports both `tags=go,k8s` (via `comma` tag) and `id=1&id=2` (repeated parameters).
- **Direct URL String Parsing**: `DecodeURL(rawURL, &filter)` parses full query strings directly.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-urlquery
```

### Complete HTTP Handler Example
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/shopsilo/go-urlquery"
)

type ProductFilter struct {
	Page     int      `query:"page,default=1"`
	Limit    int      `query:"limit,default=20"`
	Search   string   `query:"q"`
	Category []string `query:"category,comma"`
	Active   bool     `query:"active,default=true"`
}

func listProductsHandler(w http.ResponseWriter, r *http.Request) {
	var filter ProductFilter

	// Automatically decodes ?page=2&limit=50&category=electronics,books&q=laptop
	if err := urlquery.Decode(r.URL.Query(), &filter); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Search: %s | Page: %d | Limit: %d | Categories: %v\n",
		filter.Search, filter.Page, filter.Limit, filter.Category)
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestDecode_Values
--- PASS: TestDecode_Values (0.00s)
=== RUN   TestDecodeURL
--- PASS: TestDecodeURL (0.00s)
=== RUN   TestEncode
--- PASS: TestEncode (0.00s)
=== RUN   TestErrors
--- PASS: TestErrors (0.00s)
PASS
ok  	go-urlquery	0.511s
```

---

## 📄 License
MIT License.
