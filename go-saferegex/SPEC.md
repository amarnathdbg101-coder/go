# SPEC: go-saferegex

## 1. Metadata
- **Module**: `go-saferegex`
- **Package**: `saferegex`
- **Location**: `C:\Library\go-saferegex`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `regexp`, `context`, `time`, `sync`, `errors`, `fmt`).

---

## 2. Core Capabilities
1. **ReDoS Defense & Timeout Isolation**:
   - Executes regular expression evaluations inside bounded worker routines with hard timeout deadlines, shielding backend microservices from catastrophic backtracking CPU lockups.
2. **LRU Regex Compilation Cache**:
   - Pre-compiles and caches compiled `*regexp.Regexp` patterns in an $O(1)$ LRU cache to eliminate duplicate compilation latency across hot API paths.
3. **Safe String Utilities**:
   - `MatchString`, `FindString`, and `ReplaceAllString` with timeout limits and context awareness.

---

## 3. Public API
```go
type Config struct {
    Capacity       int
    DefaultTimeout time.Duration
}

type Engine struct
func New(cfg Config) *Engine
func (e *Engine) MatchString(ctx context.Context, pattern, s string, timeout time.Duration) (bool, error)
func (e *Engine) FindString(ctx context.Context, pattern, s string, timeout time.Duration) (string, error)
func (e *Engine) ReplaceAllString(ctx context.Context, pattern, src, repl string, timeout time.Duration) (string, error)

func MatchString(pattern, s string, timeout time.Duration) (bool, error)
func FindString(pattern, s string, timeout time.Duration) (string, error)
func ReplaceAllString(pattern, src, repl string, timeout time.Duration) (string, error)
```
