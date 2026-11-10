# SPEC: go-window

## 1. Scope & Purpose
`go-window` processes unbounded live event streams into discrete temporal windows (Tumbling, Sliding, Session) with watermarking and incremental reduction.

It does NOT provide persistent disk storage or distributed cluster coordination.

## 2. Public API Contract
```go
type WindowType string
type Event[T any] struct { ... }
type WindowResult[T any] struct { ... }
type Engine[T any] struct { ... }

func New[T any](cfg Config[T], opts ...Option[T]) (*Engine[T], error)
func (e *Engine[T]) Add(event Event[T]) error
func (e *Engine[T]) Flush()
func (e *Engine[T]) Close() error
```

## 3. Concurrency & Memory Guarantees
- Thread-safe via internal `sync.Mutex`.
- Expired windows trigger and purge immediately, bounding in-memory retention.
- Zero external memory leaks: completed window structures are unreferenced for GC.

## 4. How to Extend
- Add state persistence to save window progress to PostgreSQL.
- Add multi-key composite grouping.
