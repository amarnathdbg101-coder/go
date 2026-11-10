# SPEC: go-delayqueue

## 1. Scope & Purpose
`go-delayqueue` provides delayed message scheduling and execution for Go microservices. It handles future time-based task invocation, retries, and dead-letter queue isolation without requiring external tools like RabbitMQ or Celery.

It does NOT handle real-time streaming event delivery (use `go-events` or `go-pubsub`).

## 2. Public API Contract
```go
type Job struct { ... }
type Driver interface { ... }

type Queue struct { ... }
func New(driver Driver, opts ...Option) *Queue
func (q *Queue) Register(topic string, handler HandlerFunc)
func (q *Queue) Enqueue(topic string, payload []byte, delay time.Duration) (string, error)
func (q *Queue) Cancel(id string) error
func (q *Queue) Start() error
func (q *Queue) Stop() error

func EnqueueJSON[T any](q *Queue, topic string, payload T, delay time.Duration) (string, error)
```

## 3. Concurrency & Memory Guarantees
- MemoryDriver uses `container/heap` with $O(\log N)$ inserts and removals.
- Internal worker channel provides bounded backpressure without unbounded goroutine growth.
- Poller cleanly shuts down on `stopChan` close, waiting for active workers via `sync.WaitGroup`.

## 4. How to Extend
- Implement a PostgreSQL driver using `SELECT ... FOR UPDATE SKIP LOCKED` and advisory locks.
- Add metrics hooks into `go-metrics`.
