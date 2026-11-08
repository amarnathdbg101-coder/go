# Technical Specification: go-syncmap

## 1. Domain & Purpose
`go-syncmap` is a type-safe generic wrapper around Go's standard `sync.Map` (`Map[K comparable, V any]`). It completely eliminates unsafe `any` interface casting bugs and runtime assertion panics while enriching `sync.Map` with atomic `Compute` transactions, `Keys()`, `Values()`, `Len()`, and `Has()` helpers.

## 2. Architecture & Design Principles
- **Generic Compile-Time Type Safety**: Guarantees keys are `K comparable` and values are strongly-typed `V any`.
- **Full Standard Parity**: Wraps `Load`, `Store`, `LoadOrStore`, `LoadAndDelete`, `Delete`, `Swap`, `CompareAndSwap`, `CompareAndDelete`, and `Clear`.
- **Atomic Compute Support**: Implements `Compute(key, fn)` for thread-safe atomic read-modify-write workflows (e.g. counters, sets) without external lock races.
- **Convenient Inspection & Batch Iteration**: Provides `Keys()`, `Values()`, `Len()`, `Has()`, `Clone()`, and `Equal()`.
- **Zero Heavy Dependencies**: Pure standard library (`sync`, `reflect`).

## 3. API Surface
- `type Map[K comparable, V any] struct`
- `func New[K comparable, V any]() *Map[K, V]`
- `func (m *Map[K, V]) Load(key K) (V, bool)`
- `func (m *Map[K, V]) Store(key K, value V)`
- `func (m *Map[K, V]) LoadOrStore(key K, value V) (V, bool)`
- `func (m *Map[K, V]) LoadAndDelete(key K) (V, bool)`
- `func (m *Map[K, V]) Delete(key K)`
- `func (m *Map[K, V]) Swap(key K, value V) (V, bool)`
- `func (m *Map[K, V]) CompareAndSwap(key K, oldVal, newVal V) bool`
- `func (m *Map[K, V]) CompareAndDelete(key K, oldVal V) bool`
- `func (m *Map[K, V]) Range(f func(key K, value V) bool)`
- `func (m *Map[K, V]) Keys() []K`
- `func (m *Map[K, V]) Values() []V`
- `func (m *Map[K, V]) Len() int`
- `func (m *Map[K, V]) Clear()`
- `func (m *Map[K, V]) Compute(key K, fn func(oldValue V, exists bool) (newValue V, delete bool)) V`
- `func (m *Map[K, V]) Has(key K) bool`

## 4. Error Handling
- `ErrKeyNotFound`: Returned when an explicit retrieval expects an existing key.

## 5. Verification & Testing
- Unit tests verify typed storage and retrieval, atomic swap and CAS operations, range iterations, atomic compute mutations, and concurrency under 40 parallel worker routines.
