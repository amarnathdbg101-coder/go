# SPEC: go-expr

## 1. Metadata
- **Module**: `go-expr`
- **Package**: `expr`
- **Location**: `C:\Library\go-expr`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `fmt`, `strconv`, `strings`, `unicode`, `math`, `errors`).

---

## 2. Core Capabilities
1. **Sandboxed AST Evaluation Engine**:
   - Parses mathematical formulas and logical conditions into an abstract syntax tree (AST). Immune to code injection as it executes within a safe, isolated evaluation runtime.
2. **Complete Expression Syntax**:
   - Arithmetic: `+`, `-`, `*`, `/`, `%`
   - Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
   - Boolean Logic with Short-Circuiting: `&&`, `||`, `!`
   - Grouping: `( ... )`
   - Built-ins: `min(...)`, `max(...)`, `abs(...)`, `round(...)`
3. **Compilation & Multi-Goroutine Reuse**:
   - `Compile(expr)` caches tokenization and AST parsing into an immutable `*Program` capable of executing hundreds of thousands of times across concurrent goroutines without re-parsing overhead.
4. **Type-Safe Convenience Accessors**:
   - `EvalBool(expr, env)` and `EvalFloat(expr, env)` enforce expected return types.

---

## 3. Public API
```go
type Program struct
func Compile(exprStr string) (*Program, error)
func (p *Program) Eval(env map[string]any) (any, error)
func (p *Program) EvalBool(env map[string]any) (bool, error)
func (p *Program) EvalFloat(env map[string]any) (float64, error)

func Eval(exprStr string, env map[string]any) (any, error)
func EvalBool(exprStr string, env map[string]any) (bool, error)
func EvalFloat(exprStr string, env map[string]any) (float64, error)
```
