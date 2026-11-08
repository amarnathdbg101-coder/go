# Technical Specification: go-envconfig

## 1. Domain & Purpose
`go-envconfig` is a lightweight, zero-dependency environment variable unmarshaler for Go structures. It automates 12-Factor App configuration parsing directly into strongly typed structs using declarative tags (`env`, `default`, `required`, `split`, `mask`), supporting nested hierarchies and sensitive credential masking.

## 2. Architecture & Design Principles
- **Tag-Driven Reflection**: Parses struct fields annotated with `env:"NAME"`, `default:"VAL"`, `required:"true"`, `split:","`, and `mask:"true"`.
- **Comprehensive Primitive & Composite Types**: Supports `string`, integers (`int` to `int64`), unsigned integers (`uint` to `uint64`), floating point (`float32`, `float64`), boolean (`true/1/yes/on` vs `false/0/no/off`), `time.Duration`, and homogeneous slices.
- **Hierarchical Nesting & Prefixes**: Supports custom global prefixes (e.g., `APP_`) and recursively walks embedded or nested struct fields.
- **Production Observability with Secret Masking**: Provides `DumpMasked()` to safely log application startup configuration without leaking passwords, tokens, or private keys.
- **Pluggable Lookup**: Accommodates mock environment lookups (`LookupFunc`) for isolated, non-polluting unit tests.
- **Zero Heavy Dependencies**: Pure standard library (`reflect`, `strconv`, `strings`, `time`, `os`, `fmt`).

## 3. API Surface
- `type LookupFunc func(key string) (string, bool)`
- `func Process(prefix string, target any) error`
- `func ProcessWithLookup(prefix string, target any, lookup LookupFunc) error`
- `func DumpMasked(target any) map[string]string`

## 4. Error Handling
- `ErrNotPointerToStruct`: Returned when target is nil or not a pointer to a struct.
- `ErrMissingRequired`: Returned when a field marked `required:"true"` has neither an environment variable value nor a default value.
- `ErrUnsupportedType`: Returned when encountered struct field type cannot be converted.
- `ErrInvalidValue`: Returned when string cannot be parsed into the target type (e.g., invalid integer or duration).

## 5. Verification & Testing
- Unit tests verify typed conversions, slice splitting, default values, required validations, nested structs, error paths, and masked dumps.
