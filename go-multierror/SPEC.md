# SPEC: go-multierror

## 1. Metadata
- **Module**: `go-multierror`
- **Package**: `multierror`
- **Location**: `C:\Library\go-multierror`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `errors`, `fmt`, `sync`, `strings`).

---

## 2. Core Capabilities
1. **Thread-Safe Multi-Error Aggregator**:
   - Accumulates multiple non-fatal or batch worker errors concurrently using `sync.RWMutex`.
   - Auto-flattens nested `*Error` instances and drops `nil` errors.
2. **Standard Go 1.20+ Error Tree Compatibility**:
   - Implements `Unwrap() []error`, enabling standard `errors.Is` and `errors.As` to inspect the entire error tree.
3. **Idiomatic `ErrorOrNil()` Semantic**:
   - Returns typed `nil` if no errors occurred, ensuring `if err != nil` guards function correctly.
4. **Transformations & Filtering**:
   - `Filter(predicate)` and `Prefix(prefix)` for diagnostic workflows.

---

## 3. Public API
```go
type Formatter func([]error) string
func DefaultFormatter(errs []error) string

type Error struct
func New(opts ...Option) *Error
func WithFormatter(f Formatter) Option
func Append(base error, errs ...error) error
func (e *Error) Append(errs ...error)
func (e *Error) Len() int
func (e *Error) Errors() []error
func (e *Error) ErrorOrNil() error
func (e *Error) Error() string
func (e *Error) Unwrap() []error
func (e *Error) Is(target error) bool
func (e *Error) As(target any) bool
func (e *Error) Filter(fn func(error) bool) *Error
func (e *Error) Prefix(prefix string) *Error
```
