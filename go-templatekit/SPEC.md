# Technical Specification: go-templatekit

## 1. Domain & Purpose
`go-templatekit` is an ergonomic, secure template rendering engine for Go. Designed specifically for notification alerts, transactional emails, SMS templates, and dynamic text composition, it provides intuitive `{{ variable | filter }}` syntax with support for both map and struct contexts without the boilerplate of Go's standard text/template.

## 2. Architecture & Design Principles
- **Dual Context Resolution**: Resolves variable tags from `map[string]any` or via struct reflection matching field names or `json:"name"` tags.
- **Composable Filter Pipelines**: Supports chained transformations (e.g. `{{ name | trim | upper }}`) and parameter arguments (e.g. `{{ role | default:'USER' }}`).
- **Strict vs Lenient Validation**: Strict mode catches missing variables at render time (`ErrMissingVariable`), while Lenient mode provides graceful fallback defaults.
- **Extensible Registry**: Allows registering thread-safe custom filter handlers.
- **Zero Heavy Dependencies**: Pure Go standard library (`strings`, `fmt`, `reflect`, `time`).

## 3. API Surface
- `type FilterFunc func(input string, arg string) string`
- `type Engine struct`
- `func New(strict bool) *Engine`
- `func (e *Engine) RegisterFilter(name string, fn FilterFunc)`
- `func (e *Engine) Render(tmpl string, data any) (string, error)`

## 4. Error Handling
- `ErrMissingVariable`: Variable not present in strict mode.
- `ErrUnknownFilter`: Filter name not registered in strict mode.

## 5. Verification & Testing
- Unit tests verify map and struct substitution, built-in filter pipelines (upper, lower, default, trim), strict mode enforcement, and custom filter extensions.
