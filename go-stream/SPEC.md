# SPEC: go-stream

## 1. Metadata
- **Module**: `go-stream`
- **Package**: `gostream`
- **Location**: `C:\Library\go-stream`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `sort`).

---

## 2. Core Capabilities
1. **Generic Type-Safe Pipeline**:
   - `Stream[T any]` provides fluent chaining (`Filter`, `Distinct`, `Sort`, `Take`, `Skip`, `Reduce`).
2. **Transformations Across Types**:
   - `Map[T, R]` transforms elements from type `T` to type `R`.
   - `FlatMap[T, R]` maps and flattens nested slices into a single stream.
   - `Chunk[T]` breaks a stream into fixed-size batches for bulk operations.
3. **Concurrent Order-Preserving Worker Pool**:
   - `ParallelMap[T, R](s, workers, fn)` processes elements concurrently across goroutines while guaranteeing the resulting slice preserves the original input order.

---

## 3. Public API
```go
type Stream[T any] struct { ... }

func FromSlice[T any](items []T) *Stream[T]
func Empty[T any]() *Stream[T]

func (s *Stream[T]) Filter(predicate func(T) bool) *Stream[T]
func (s *Stream[T]) Distinct(keySelector func(T) any) *Stream[T]
func (s *Stream[T]) Sort(less func(a, b T) bool) *Stream[T]
func (s *Stream[T]) Take(n int) *Stream[T]
func (s *Stream[T]) Skip(n int) *Stream[T]
func (s *Stream[T]) ToSlice() []T
func (s *Stream[T]) ForEach(action func(T))
func (s *Stream[T]) First() (T, bool)
func (s *Stream[T]) Count() int
func (s *Stream[T]) AnyMatch(predicate func(T) bool) bool
func (s *Stream[T]) AllMatch(predicate func(T) bool) bool
func (s *Stream[T]) Reduce(initial T, accumulator func(acc, item T) T) T

func Map[T, R any](s *Stream[T], transform func(T) R) *Stream[R]
func FlatMap[T, R any](s *Stream[T], transform func(T) []R) *Stream[R]
func Chunk[T any](s *Stream[T], size int) *Stream[[]T]
func ParallelMap[T, R any](s *Stream[T], workers int, transform func(T) R) *Stream[R]
```
