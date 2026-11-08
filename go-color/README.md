# go-color

Lightweight ANSI terminal styling, 24-bit TrueColor (RGB/HEX), and NO_COLOR standard detection for Go.

---

## Features
- **ANSI & 24-Bit TrueColor**: Full support for standard 16 terminal colors, custom RGB `(r, g, b)`, and hex codes (`#FF5733`).
- **NO_COLOR Compliant**: Automatically detects `NO_COLOR` standard and disables escape sequences when redirected to files or pipes.
- **Fluent Chaining**: Build complex styles effortlessly: `color.New().Bold().Red().BgYellow().Sprint("Alert")`.
- **Zero Allocations**: Instant string building without third-party dependencies.

---

## Quickstart

```go
package main

import (
	"fmt"

	gocolor "go-color"
)

func main() {
	// 1. Quick helpers
	fmt.Println(gocolor.Green("✔ Service connected successfully!"))
	fmt.Println(gocolor.Red("✖ Database connection failed!"))
	fmt.Println(gocolor.Yellow("⚠ High memory utilization warning!"))

	// 2. Chained styling with TrueColor HEX
	brandStyle, _ := gocolor.New().Bold().HEX("#FF8C00")
	fmt.Println(brandStyle.Sprint("ShopSilo Enterprise Backend"))
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
