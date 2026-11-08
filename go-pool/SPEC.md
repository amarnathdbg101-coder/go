# Technical Specification: go-pool

## 1. Domain & Purpose
`go-pool` is a bounded, generic object and connection pool for Go. Unlike standard `sync.Pool` (which clears completely on GC cycles, cannot be bounded, and cannot health check objects), `go-pool` manages long-lived, high-cost resources (such as TCP/WebSocket sockets, database sessions, heavy buffers, or worker processes) with capacity capping, context-aware acquisition, idle eviction, and automatic disposal.

## 2. Architecture & Design Principles
- **Strict Capacity Bounding**: Regulates concurrent live resources using internal semaphore buffering to prevent system resource exhaustion.
- **Context-Aware Acquisition**: `Acquire(ctx)` honors deadline deadlines and cancellation without blocking caller goroutines permanently.
- **Active Health Checks**: Validates object integrity prior to checkout via `HealthCheck` callback, automatically discarding and replacing defunct instances.
- **Idle Timeout Eviction**: Prunes stale idle connections exceeding configured maximum idle age.
- **Zero Heavy Dependencies**: Pure Go standard library (`context`, `sync`, `time`, `io`).

## 3. API Surface
- `type Config[T any] struct { Capacity int; Factory func(ctx context.Context) (T, error); CloseFn func(item T) error; HealthCheck func(item T) bool; IdleTimeout time.Duration }`
- `type Pool[T any] struct`
- `func New[T any](cfg Config[T]) (*Pool[T], error)`
- `func (p *Pool[T]) Acquire(ctx context.Context) (T, error)`
- `func (p *Pool[T]) Release(item T)`
- `func (p *Pool[T]) Discard(item T)`
- `func (p *Pool[T]) Idle() int`
- `func (p *Pool[T]) Active() int`
- `func (p *Pool[T]) Close() error`

## 4. Error Handling
- `ErrPoolClosed`: Operation on closed pool.
- `ErrPoolExhausted`: Maximum capacity exhausted.
- `ErrInvalidConfig`: Capacity <= 0 or nil Factory.

## 5. Verification & Testing
- Unit tests verify instance reuse across Acquire/Release cycles, capacity throttling with deadline timeouts, health-check automatic discarding, and idle timeout pruning.
