# go-colorgrad 🎨

> **24-bit TrueColor Terminal Gradient Engine & RGB Interpolator for Go (Go 1.23+).**  
> CLI tools, logs, dashboards, aur terminal applications ke liye smooth color transitions, gradient text styling, aur dynamic progress bars.

---

## 🚀 Features

- **24-Bit ANSI TrueColor**: Modern terminals me 16 million colors (`\x1b[38;2;R;G;Bm`) support karta hai.
- **RGB Interpolation**: Do ya do se zyada color stops ke beech linear lerp calculation.
- **Dynamic Progress Bar**: Percentage ke according color badalne wala progress bar widget (`TrafficLight` preset se Red -> Yellow -> Green).
- **NO_COLOR Compliant**: Agar environment me `NO_COLOR=1` set ho, toh plain text return hota hai.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-colorgrad"
)

func main() {
	// 1. Neon Gradient Text
	neon := colorgrad.Neon()
	fmt.Println(neon.Apply("🚀 Antigravity High Performance Engine"))

	// 2. Dynamic Progress Bar (75% complete)
	traffic := colorgrad.TrafficLight()
	fmt.Println(colorgrad.ProgressBar(0.75, 25, traffic))
}
```
