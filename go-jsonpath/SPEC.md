# Technical Specification: go-jsonpath

## 1. Domain & Purpose
`go-jsonpath` is a lightweight, zero-dependency JSONPath parser and extractor for Go. It allows extracting arbitrary values, nested objects, array elements, and multi-element projections (`[*]`) directly from JSON payloads or in-memory maps without requiring pre-declared Go struct definitions.

## 2. Architecture & Design Principles
- **Ergonomic Path Expressions**: Supports dot notation (`$.store.name`), array indices (`$.users[0]`), and wildcard array projections (`$.items[*].id`).
- **Optional Root Anchor**: Handles expressions with or without the leading `$` or `$.`.
- **Type-Safe Extractors**: Provides specialized helpers (`GetString`, `GetInt`, `GetFloat`, `GetBool`, `GetSlice`) with automatic numeric casting from JSON `float64`.
- **Informative Error Model**: Distinct sentinel errors for syntax parsing errors, missing keys, array bounds, invalid JSON, and type mismatches.
- **Zero Heavy Dependencies**: Pure standard library (`encoding/json`, `strings`, `strconv`, `fmt`).

## 3. API Surface
- `func ParsePath(path string) ([]token, error)`
- `func Get(jsonData []byte, path string) (any, error)`
- `func GetFromAny(data any, path string) (any, error)`
- `func GetString(jsonData []byte, path string) (string, error)`
- `func GetInt(jsonData []byte, path string) (int64, error)`
- `func GetFloat(jsonData []byte, path string) (float64, error)`
- `func GetBool(jsonData []byte, path string) (bool, error)`
- `func GetSlice(jsonData []byte, path string) ([]any, error)`

## 4. Error Handling
- `ErrInvalidPath`: Syntax errors (e.g. unclosed brackets, non-integer index).
- `ErrNotFound`: Requested field key does not exist.
- `ErrIndexOutOfBounds`: Requested array index is negative or >= slice length.
- `ErrTypeMismatch`: Value exists but cannot be converted to target type.
- `ErrInvalidJSON`: Malformed input JSON payload.

## 5. Verification & Testing
- Unit tests verify nested object traversal, numeric and boolean extraction, array indexing, wildcard slice flattening, and strict error reporting.
