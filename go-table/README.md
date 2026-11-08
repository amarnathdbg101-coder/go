# go-table 📑

`go-table` ek lightweight, zero-dependency ASCII grid, Markdown, aur CSV table formatting library hai Go ke liye. Ye terminal output, CLI tools, aur reports ke liye structured tabular data ko beautifully aligned tables me render karta hai with Unicode width calculation aur custom column justification (Left, Center, Right).

---

## ✨ Features

- **Multiple Render Formats**:
  - `Render()`: Clean ASCII boxed terminal grid (`+---+---+`).
  - `RenderMarkdown()`: GitHub-flavored Markdown table with alignment indicators (`:---`, `:---:`, `---:`).
  - `RenderCSV()`: Standard comma-separated output with automatic escaping.
- **Unicode Width Alignment**: Multilingual characters (e.g. Devanagari/Hindi) border alignment ko distort nahi karte.
- **Column Justification**: `AlignLeft`, `AlignCenter`, `AlignRight`.
- **Dynamic Cell Formatting**: Pass any primitives (int, float, string, boolean).
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-table
```

### Complete Example
```go
package main

import (
	"fmt"

	"github.com/shopsilo/go-table"
)

func main() {
	tbl := table.New("ID", "Customer", "Amount (₹)", "Status")
	tbl.AddRow(1, "Rahul Sharma", 4500.00, "PAID")
	tbl.AddRow(2, "Sneha Patel", 1250.50, "PENDING")
	tbl.AddRow(3, "Amit Verma", 8900.00, "PAID")

	// Justify columns: ID right, Amount right, Status center
	tbl.SetAlignment(0, table.AlignRight)
	tbl.SetAlignment(2, table.AlignRight)
	tbl.SetAlignment(3, table.AlignCenter)

	// 1. Terminal ASCII Output
	fmt.Println("--- Terminal Output ---")
	fmt.Println(tbl.Render())

	// 2. Markdown Output
	fmt.Println("\n--- Markdown Output ---")
	fmt.Println(tbl.RenderMarkdown())
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestTable_RenderASCII
--- PASS: TestTable_RenderASCII (0.00s)
=== RUN   TestTable_RenderMarkdown
--- PASS: TestTable_RenderMarkdown (0.00s)
=== RUN   TestTable_RenderCSV
--- PASS: TestTable_RenderCSV (0.00s)
=== RUN   TestTable_Unicode
--- PASS: TestTable_Unicode (0.00s)
PASS
ok  	go-table	0.475s
```

---

## 📄 License
MIT License.
