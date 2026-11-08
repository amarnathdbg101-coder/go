# Technical Specification: go-batcher

## 1. Domain & Purpose
`go-batcher` is a high-throughput, thread-safe micro-batching aggregator for Go. In high-traffic systems, executing database writes, event emissions, or API calls individually degrades throughput due to round-trip latency and connection overhead. `go-batcher` buffers items concurrently and dispatches them in batches either when size limits are reached or when a flush deadline elapses.

## 2. Architecture & Design Principles
- **Dual-Trigger Flushing**: Triggers flush either on item count threshold (`MaxBatchSize`) or timeout tick (`FlushInterval`), whichever arrives first.
- **Zero-Loss Graceful Teardown**: `Stop()` signals worker shutdown and completely drains and flushes all buffered items before unblocking.
- **Backpressure & Saturation Control**: Provides both blocking `Push(item)` and non-blocking `TryPush(item)` returning `ErrQueueFull`.
- **Panic & Failure Isolation**: Dedicated `OnPanic` and `OnError` callbacks prevent worker crashes and record batch-level dispatch failures.
- **Zero Heavy Dependencies**: Pure Go standard library (`sync`, `time`, `fmt`).

## 3. API Surface
- `type FlushFunc[T any] func(batch []T) error`
- `type Config[T any] struct { MaxBatchSize, QueueCapacity int; FlushInterval time.Duration; FlushFn FlushFunc[T]; OnError func(batch []T, err error); OnPanic func(r any) }`
- `func New[T any](cfg Config[T]) (*Batcher[T], error)`
- `func (b *Batcher[T]) Push(item T) error`
- `func (b *Batcher[T]) TryPush(item T) error`
- `func (b *Batcher[T]) Flush()`
- `func (b *Batcher[T]) Stop()`

## 4. Error Handling
- `ErrBatcherStopped`: Pushing to a stopped batcher.
- `ErrQueueFull`: Non-blocking push on a saturated queue.
- `ErrInvalidConfig`: Invalid config options.

## 5. Verification & Testing
- Unit tests verify threshold flushes by count, timer-based flushes, manual synchronous flushes, non-blocking backpressure rejection, zero-loss graceful stop, and panic isolation.
