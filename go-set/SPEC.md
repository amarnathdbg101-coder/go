# Technical Specification: go-set

## 1. Domain & Purpose
`go-set` is a high-performance, type-safe generic `Set[T comparable]` data structure for Go. It eliminates boilerplate `map[T]struct{}` code by providing set algebra operations (Union, Intersect, Difference, Symmetric Difference, Subsets/Supersets), functional filtering, JSON serialization, and optional thread safety.

## 2. Architecture & Design Principles
- **Generic Element Type**: Works with any Go comparable type (`string`, integers, floats, pointers, comparable structs).
- **Dual Concurrency Modes**: Supports lightweight, lock-free instances (`New()`) for local scope and RWMutex-protected instances (`NewThreadSafe()`) for concurrent sharing.
- **Mathematical Set Algebra**: Full support for Union, Intersection, Set Difference ($A \setminus B$), Symmetric Difference ($A \Delta B$), Subset, and Superset.
- **Native JSON Serialization**: Implements `json.Marshaler` and `json.Unmarshaler` for API JSON payloads.
- **Zero Heavy Dependencies**: Pure standard library (`sync`, `encoding/json`).

## 3. API Surface
- `type Set[T comparable] struct`
- `func New[T comparable](items ...T) *Set[T]`
- `func NewThreadSafe[T comparable](items ...T) *Set[T]`
- `func (s *Set[T]) Add(items ...T)`
- `func (s *Set[T]) Remove(items ...T)`
- `func (s *Set[T]) Contains(item T) bool`
- `func (s *Set[T]) ContainsAll(items ...T) bool`
- `func (s *Set[T]) ContainsAny(items ...T) bool`
- `func (s *Set[T]) Size() int`
- `func (s *Set[T]) IsEmpty() bool`
- `func (s *Set[T]) Clear()`
- `func (s *Set[T]) ToSlice() []T`
- `func (s *Set[T]) Clone() *Set[T]`
- `func (s *Set[T]) Equal(other *Set[T]) bool`
- `func (s *Set[T]) Union(other *Set[T]) *Set[T]`
- `func (s *Set[T]) Intersect(other *Set[T]) *Set[T]`
- `func (s *Set[T]) Difference(other *Set[T]) *Set[T]`
- `func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T]`
- `func (s *Set[T]) IsSubset(other *Set[T]) bool`
- `func (s *Set[T]) IsSuperset(other *Set[T]) bool`
- `func (s *Set[T]) Filter(predicate func(T) bool) *Set[T]`
- `func (s *Set[T]) Each(fn func(T))`

## 4. Error Handling
- `ErrNilSet`: Indicates an uninitialized or nil set operand.

## 5. Verification & Testing
- Unit tests verify core set operations, algebraic set equations, filter transformations, JSON roundtrip marshaling, and concurrent thread-safe mutations.
