# Technical Specification: go-strcase

## 1. Domain & Purpose
`go-strcase` is a high-performance, minimal-allocation string casing conversion engine for Go. It converts identifiers, field names, and tokens between common casing styles including `snake_case`, `camelCase`, `PascalCase`, `kebab-case`, and `SCREAMING_SNAKE_CASE`. It features built-in handling of continuous uppercase acronyms (e.g. `HTTPServer` -> `http_server`, `userID` -> `user_id`) without external dependencies.

## 2. Architecture & Design Principles
- **Smart Acronym Splitting**: Distinguishes between multi-character uppercase acronyms and lowercase transition boundaries (`HTTPServer` splits cleanly to `["HTTP", "Server"]`, rather than `["H", "T", "T", "P", "Server"]`).
- **Delimiter Agnostic**: Correctly parses underscores, hyphens, spaces, dots, slashes, and mixed inputs.
- **Allocation Efficiency**: Leverages `strings.Builder` and direct rune buffers for fast string reconstructions.
- **Zero Heavy Dependencies**: Pure Go standard library (`strings`, `unicode`).

## 3. API Surface
- `func SplitWords(s string) []string`
- `func ToSnake(s string) string`
- `func ToCamel(s string) string`
- `func ToPascal(s string) string`
- `func ToKebab(s string) string`
- `func ToScreamingSnake(s string) string`
- `func ToCustom(s string, delimiter rune, uppercase bool) string`

## 4. Verification & Testing
- Unit tests verify word splitting across snake_case, kebab-case, camelCase, PascalCase, uppercase acronyms, URL paths, dotted namespaces, and empty/single-character inputs.
