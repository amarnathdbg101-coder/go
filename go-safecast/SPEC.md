# Technical Specification: go-safecast

## 1. Domain & Purpose
`go-safecast` is a zero-allocation, type-safe numeric casting library for Go. Go's native type casting silently truncates and overflows integers and floats (for instance, `int8(300)` results in `44`, and `uint(int(-1))` wraps to `MaxUint`), causing silent data corruption, security exploits, or financial inaccuracies. `go-safecast` uses Go generics to detect overflows, underflows, and invalid float representations at runtime.

## 2. Architecture & Design Principles
- **Generics-Driven**: Defines `Number`, `Integer`, `Signed`, `Unsigned`, and `Float` type constraints.
- **Strict Bounds Verification**: Checks mathematical min/max limits before downcasting.
- **Float & Special Values Guarding**: Intercepts IEEE-754 `NaN` (Not a Number) and `+Inf`/`-Inf` (Infinities).
- **Zero Allocations**: Pure value conversions with 0 heap allocations.
- **Zero External Dependencies**: Pure Go standard library (`math`, `errors`).

## 3. API Surface
- `func ToInt[T Number](v T) (int, error)`
- `func ToInt8[T Number](v T) (int8, error)`
- `func ToInt16[T Number](v T) (int16, error)`
- `func ToInt32[T Number](v T) (int32, error)`
- `func ToInt64[T Number](v T) (int64, error)`
- `func ToUint[T Number](v T) (uint, error)`
- `func ToUint8[T Number](v T) (uint8, error)`
- `func ToUint16[T Number](v T) (uint16, error)`
- `func ToUint32[T Number](v T) (uint32, error)`
- `func ToUint64[T Number](v T) (uint64, error)`
- `func Must[T any](v T, err error) T`

## 4. Error Handling
- `ErrOverflow`: Value exceeds the maximum representable limit of destination type.
- `ErrUnderflow`: Value is below minimum representable limit (or negative into unsigned).
- `ErrNonInteger`: Value has fractional portion when integer is required.

## 5. Verification & Testing
- Unit tests verify signed and unsigned boundaries, 32-bit and 64-bit limits, negative-to-unsigned conversion prevention, float NaN/Inf rejection, and panic behavior of the `Must` helper.
