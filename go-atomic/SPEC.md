# Technical Specification: go-atomic

## 1. Domain & Purpose
`go-atomic` is an ergonomic, type-safe suite of high-level atomic primitives for Go. It provides lock-free, race-free wrappers for Booleans (with atomic toggling), Strings, Durations, Times, IEEE-754 Floats (with CAS-loop math), and Generic Struct Values.

## 2. Architecture & Design Principles
- **Ergonomic Memory Safety**: Eliminates the unsafe pointer arithmetic of raw `sync/atomic` calls.
- **Atomic Operations**: Full support for `.Load()`, `.Store()`, `.Swap()`, `.CAS()`, and `.Toggle()`.
- **Atomic Float64 Math**: Uses IEEE 754 bit-pattern casting (`math.Float64bits` / `math.Float64frombits`) with an optimized Compare-And-Swap retry loop.
- **Type-Safe Generic Value Container**: `Value[T]` wraps `atomic.Value` with zero type assertion panics.
- **JSON Serialization**: Direct JSON marshaling/unmarshaling for configuration flags.
- **Zero Heavy Dependencies**: Pure standard library (`sync/atomic`, `math`, `time`, `encoding/json`).

## 3. API Surface
- `type Bool struct`
- `func NewBool(initial bool) *Bool`
- `func (b *Bool) Load() bool`
- `func (b *Bool) Store(val bool)`
- `func (b *Bool) Swap(val bool) bool`
- `func (b *Bool) CAS(oldVal, newVal bool) bool`
- `func (b *Bool) Toggle() bool`
- `type String struct`
- `func NewString(initial string) *String`
- `func (s *String) Load() string`
- `func (s *String) Store(val string)`
- `func (s *String) Swap(val string) string`
- `type Duration struct`
- `func NewDuration(initial time.Duration) *Duration`
- `func (d *Duration) Load() time.Duration`
- `func (d *Duration) Store(val time.Duration)`
- `func (d *Duration) Add(delta time.Duration) time.Duration`
- `type Time struct`
- `func NewTime(initial time.Time) *Time`
- `func (t *Time) Load() time.Time`
- `func (t *Time) Store(val time.Time)`
- `type Float64 struct`
- `func NewFloat64(initial float64) *Float64`
- `func (f *Float64) Load() float64`
- `func (f *Float64) Store(val float64)`
- `func (f *Float64) Add(delta float64) float64`
- `func (f *Float64) CAS(oldVal, newVal float64) bool`
- `type Value[T any] struct`
- `func NewValue[T any](initial T) *Value[T]`
- `func (val *Value[T]) Load() T`
- `func (val *Value[T]) Store(newVal T)`
- `func (val *Value[T]) Swap(newVal T) T`

## 4. Error Handling
- `ErrNilValue`: Disallowed attempt to store a nil pointer into an atomic typed container.

## 5. Verification & Testing
- Unit tests verify atomic swapping, CAS transitions, toggle logic, duration arithmetic, float CAS addition, generic struct safety, and concurrent race-free execution under 40 parallel goroutines.
