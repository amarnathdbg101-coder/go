# Technical Specification: go-urlquery

## 1. Domain & Purpose
`go-urlquery` is an ergonomic URL query parameter encoder and decoder for Go. It allows direct, strongly-typed unmarshaling from `url.Values` or raw request URLs into Go structs, supporting default values, comma-delimited slices, repeated array keys, and durations without manual conversion boilerplate.

## 2. Architecture & Design Principles
- **Declarative Struct Tags**:
  - `query:"name"`: Sets parameter key name.
  - `default=value`: Fallback value applied if query parameter is absent or empty.
  - `comma`: Parses comma-separated values (`tags=go,k8s`) into slice fields.
  - `-`: Excludes fields from serialization.
- **Bi-Directional Serialization**: Full support for both `Decode()` (`url.Values` $\to$ struct) and `Encode()` (struct $\to$ `url.Values` and query strings).
- **Comprehensive Primitive Support**: Decodes `string`, integers, unsigned integers, floats, booleans, `time.Duration`, and slices.
- **Zero Heavy Dependencies**: Pure standard library (`net/url`, `reflect`, `strconv`, `strings`, `time`).

## 3. API Surface
- `func Decode(values url.Values, target any) error`
- `func DecodeURL(rawURL string, target any) error`
- `func Encode(source any) (url.Values, error)`
- `func EncodeString(source any) (string, error)`

## 4. Error Handling
- `ErrNotPointerToStruct`: Target must be a non-nil pointer to a struct.
- `ErrNotStruct`: Source for encoding must be a struct.
- `ErrUnsupportedType`: Encountered struct field type cannot be converted.
- `ErrInvalidValue`: Value string cannot be parsed into target type.

## 5. Verification & Testing
- Unit tests verify typed decoding, default fallback injection, comma-delimited and multi-value slices, raw URL parsing, bidirectional encoding, and type validation errors.
