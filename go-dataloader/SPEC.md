# SPEC: go-dataloader

## 1. Scope & Purpose
`go-dataloader` provides a generic request batching and caching engine for Go applications. It is modeled after the DataLoader pattern (Facebook) to solve the $N+1$ read query problem in distributed microservices and relational databases without external services or heavy dependencies.

It explicitly does NOT manage database connections or persistence layers itself.

## 2. Public API Contract
```go
type BatchFunc[K comparable, V any] func(ctx context.Context, keys []K) ([]Result[V], error)

type Loader[K comparable, V any] struct { ... }

func New[K comparable, V any](batchFn BatchFunc[K, V], opts ...Option) *Loader[K, V]
func (l *Loader[K, V]) Load(ctx context.Context, key K) (V, error)
func (l *Loader[K, V]) LoadMany(ctx context.Context, keys []K) ([]V, []error)
func (l *Loader[K, V]) Prime(key K, value V)
func (l *Loader[K, V]) Clear(key K)
func (l *Loader[K, V]) ClearAll()
func (l *Loader[K, V]) Close() error
```

## 3. Concurrency & Memory Guarantees
- Thread-safe via internal `sync.Mutex`.
- Single batch allocation per tick cycle.
- Timer automatically stops and drains when closed or dispatched.
- Goroutine isolation: user's `batchFn` executes outside locks to prevent contention.

## 4. How to Extend
- Implement custom `BatchFromMap` or RPC client adapters.
- Use `WithWait` and `WithMaxBatch` to tune latency vs throughput for high-QPS services.
