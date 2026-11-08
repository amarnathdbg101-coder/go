# go-templatekit 📝

`go-templatekit` ek lightweight, high-performance aur secure template rendering engine hai Go ke liye. Ye transactional emails, SMS notifications, dynamic text composition, aur message formatting ke liye banaya gaya hai jisme intuitive `{{ variable | filter }}` syntax milta hai bina Go standard `text/template` ke heavy boilerplate ke.

---

## ✨ Features

- **Intuitive Variable Substitution**: Direct `{{ variable }}` placeholder replacement.
- **Dual Context Support**: Dono `map[string]any` aur Go `struct` (field names ya `json` struct tags) seamlessly support karta hai.
- **Chained Filter Pipelines**: Unix-style composable filters jaise `{{ name | trim | upper }}` ya `{{ role | default:'USER' }}`.
- **Built-in Rich Filters**: `upper`, `lower`, `trim`, `title`, `default`.
- **Strict & Lenient Modes**: Strict mode me missing variables ya unknown filters par instant error milti hai (`ErrMissingVariable`), jabki lenient mode me graceful defaults aate hain.
- **Thread-Safe Custom Filters**: Asaani se apne custom business filter functions register karein (`RegisterFilter`).
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-templatekit
```

### Basic Usage with Map Context
```go
package main

import (
	"fmt"
	"log"

	"github.com/shopsilo/go-templatekit"
)

func main() {
	engine := templatekit.New(true) // strict mode

	tmpl := "Hello {{ user.name | trim | upper }}! Your balance is {{ balance }}."
	data := map[string]any{
		"user.name": "  rahul sharma  ",
		"balance":   5000,
	}

	result, err := engine.Render(tmpl, data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
	// Output: Hello RAHUL SHARMA! Your balance is 5000.
}
```

### Struct Support & Default Values
```go
type OrderNotification struct {
	OrderID string  `json:"order_id"`
	Status  string  `json:"status"`
	Courier string  `json:"courier"`
}

func main() {
	engine := templatekit.New(false) // lenient mode

	tmpl := "Order #{{ order_id }}: Status is {{ status | upper }}. Shipped via {{ courier | default:'Standard Post' }}."
	order := OrderNotification{
		OrderID: "ORD-9921",
		Status:  "dispatched",
		Courier: "", // empty will trigger default filter
	}

	res, _ := engine.Render(tmpl, order)
	fmt.Println(res)
	// Output: Order #ORD-9921: Status is DISPATCHED. Shipped via Standard Post.
}
```

### Registering Custom Filter
```go
engine := templatekit.New(true)

// Mask filter: replaces all chars except first 2 with '*'
engine.RegisterFilter("mask", func(input string, arg string) string {
    if len(input) <= 2 {
        return input
    }
    return input[:2] + "****"
})

tmpl := "Card: {{ card_no | mask }}"
res, _ := engine.Render(tmpl, map[string]any{"card_no": "12345678"})
// Output: Card: 12****
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestEngine_Render_Map
--- PASS: TestEngine_Render_Map (0.00s)
=== RUN   TestEngine_Render_Struct
--- PASS: TestEngine_Render_Struct (0.00s)
=== RUN   TestEngine_Render_Filters
--- PASS: TestEngine_Render_Filters (0.00s)
=== RUN   TestEngine_Render_StrictError
--- PASS: TestEngine_Render_StrictError (0.00s)
=== RUN   TestEngine_CustomFilter
--- PASS: TestEngine_CustomFilter (0.00s)
PASS
ok  	github.com/shopsilo/go-templatekit	0.479s
```

---

## 📄 License
MIT License.
