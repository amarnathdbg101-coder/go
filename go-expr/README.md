# go-expr 🧮

> **Safe, Lightweight In-Memory Formula & Expression Evaluator for Go (Go 1.23+).**  
> Dynamic business rules, discount calculations, feature conditions, aur formula evaluation ke liye safe runtime engine. Bina kisi external scripting engine (Lua/JS) ke pure Go AST me evaluate hota hai, isliye code injection attacks impossible hain.

---

## 🚀 Features

- **Sandboxed Execution**: Pure recursive descent AST interpreter. Host machine ya Go runtime me koi unauthorized memory access nahi ho sakta.
- **Full Expression Grammar**:
  - Arithmetic: `+`, `-`, `*`, `/`, `%`
  - Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
  - Boolean Logic: `&&`, `||`, `!` (Short-circuiting support)
  - Parentheses: `( ... )`
  - Functions: `min()`, `max()`, `abs()`, `round()`
- **High Throughput**: `Compile()` se AST ek baar build hota hai aur millions of concurrent requests me reuse kiya ja sakta hai.

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-expr"
)

func main() {
	// 1. Dynamic Discount Formula
	formula := "price * quantity * (1 - discount)"
	env := map[string]any{
		"price":    250.0,
		"quantity": 4,
		"discount": 0.15, // 15% off
	}

	total, _ := expr.EvalFloat(formula, env)
	fmt.Printf("Total Amount: ₹%.2f\n", total) // Output: ₹850.00

	// 2. Pre-compiled Gating Rule
	rule, _ := expr.Compile("age >= 18 && (role == 'admin' || points > 100)")
	user := map[string]any{"age": 22, "role": "member", "points": 150}

	canAccess, _ := rule.EvalBool(user)
	fmt.Printf("Access Granted: %v\n", canAccess) // Output: true
}
```
