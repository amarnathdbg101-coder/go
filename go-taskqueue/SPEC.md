# SPEC: go-taskqueue

## 1. Scope & Purpose
Production-ready asynchronous background task execution queue with bounded worker pools, automatic exponential backoff retries, panic recovery, and graceful shutdown.
- In Scope: In-Memory channel queue, bounded worker pool concurrency, per-task execution timeouts, exponential retry backoff, panic safety, typed JSON payload unmarshaling, and graceful shutdown waiting for active tasks.
- Out of Scope: Multi-datacenter message brokers (Kafka/RabbitMQ clustering).

## 2. Public API Contract
- `New(opts ...Option) *Engine`
- `NewWithQueue(q Queue, opts ...Option) *Engine`
- `Engine.Register(taskType string, handler HandlerFunc)`
- `Engine.Start()`
- `Engine.Enqueue(taskType string, payload any) (*Task, error)`
- `Engine.EnqueueWait(ctx context.Context, taskType string, payload any) (*Task, error)`
- `Engine.OnFailure(fn func(task *Task, err error))`
- `Engine.Shutdown(ctx context.Context) error`
- `Task.UnmarshalPayload(dst any) error`
- `Queue` interface: `Enqueue`, `EnqueueWait`, `Dequeue`, `Len`, `Close`.

## 3. Concurrency & Memory Guarantees
- Bounded Workers: Uses a fixed number of goroutines (`cfg.Workers`), eliminating unbounded goroutine leaks.
- Zero Process Crashes: Each worker has isolated `defer recover()` protection. A task panic does not crash the server.
- Graceful Shutdown: `Shutdown(ctx)` closes the queue channel and waits for active workers via `sync.WaitGroup` before terminating.

## 4. How to Extend
- Implement the `Queue` interface (`Enqueue`, `EnqueueWait`, `Dequeue`, `Len`, `Close`) to plug in Redis (LPUSH/BRPOP or Streams), PostgreSQL LISTEN/NOTIFY, or AWS SQS.
