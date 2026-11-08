# go-jsonpath 🔍

`go-jsonpath` ek lightweight, zero-dependency JSONPath parser aur extraction library hai Go ke liye. Ye arbitrary JSON payloads, dynamic API responses, aur webhooks se specific fields ya array slices extract karta hai bina pre-defined Go structs likhe.

---

## ✨ Features

- **Ergonomic Path Syntax**: `$.user.profile.name`, `items[0].id`, `catalog.books[*].price`.
- **Flexible Anchor**: Leading `$` prefix optional hai (`$.a.b` aur `a.b` dono supported).
- **Wildcard Projections**: `[*]` array iteration se children properties ko flatten slice me nikalta hai.
- **Strongly Typed Getters**: `GetString()`, `GetInt()`, `GetFloat()`, `GetBool()`, `GetSlice()`.
- **Smart Numeric Casting**: JSON standard `float64` numbers ko safely `int64` ya `float64` me convert karta hai.
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-jsonpath
```

### Complete Example
```go
package main

import (
	"fmt"
	"log"

	"github.com/shopsilo/go-jsonpath"
)

func main() {
	rawJSON := []byte(`{
		"order": {
			"id": "ORD-101",
			"paid": true,
			"total": 1250.50,
			"items": [
				{"name": "Wireless Mouse", "qty": 2, "price": 450},
				{"name": "Mechanical Keyboard", "qty": 1, "price": 800.5}
			]
		}
	}`)

	// Extract nested string
	orderID, _ := jsonpath.GetString(rawJSON, "$.order.id")
	fmt.Printf("Order: %s\n", orderID)

	// Extract nested boolean & float
	isPaid, _ := jsonpath.GetBool(rawJSON, "order.paid")
	total, _ := jsonpath.GetFloat(rawJSON, "order.total")
	fmt.Printf("Paid: %v, Total: ₹%.2f\n", isPaid, total)

	// Extract array item by index
	firstItemName, _ := jsonpath.GetString(rawJSON, "$.order.items[0].name")
	fmt.Printf("First item: %s\n", firstItemName)

	// Extract array wildcard projection
	allItemNames, err := jsonpath.GetSlice(rawJSON, "$.order.items[*].name")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("All items: %v\n", allItemNames)
	// Output: All items: [Wireless Mouse Mechanical Keyboard]
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestGet_BasicFields
--- PASS: TestGet_BasicFields (0.00s)
=== RUN   TestGet_ArrayIndexing
--- PASS: TestGet_ArrayIndexing (0.00s)
=== RUN   TestGet_WildcardProjection
--- PASS: TestGet_WildcardProjection (0.00s)
=== RUN   TestGet_Errors
--- PASS: TestGet_Errors (0.00s)
PASS
ok  	go-jsonpath	0.528s
```

---

## 📄 License
MIT License.
