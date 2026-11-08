# Technical Specification: go-semaphore

## 1. Domain & Purpose
`go-semaphore` is a weighted counting semaphore with FIFO fairness, context cancellation support, non-blocking polling (`TryAcquire`), timeout handling (`AcquireTimeout`), and graceful draining. It is designed to govern concurrent workloads, rate-limit outbound network calls, and control database connection pressure.

## 2. Architecture & Design Principles
- **Weighted Token Acquisition**: Supports acquiring and releasing arbitrary weights ($n > 0$).
- **Strict FIFO Fairness**: Queued requests are prioritized in arrival order to prevent thread starvation under heavy contention.
- **Leak-Proof Context Cancellation**: Waiters cleanly deregister on `ctx.Done()`; if tokens were granted concurrently during cancellation, they are refunded immediately to unblock downstream waiters.
- **Graceful Lifecycle & Draining**: Supports non-blocking token draining (`Drain()`) and atomic closure (`Close()`), signaling waiting goroutines with `ErrClosed`.
- **Zero Heavy Dependencies**: Pure standard library (`sync`, `context`, `time`, `errors`).

## 3. API Surface
- `type Semaphore struct`
- `func New(capacity int64) *Semaphore`
- `func (s *Semaphore) Acquire(ctx context.Context, n int64) error`
- `func (s *Semaphore) AcquireTimeout(n int64, d time.Duration) error`
- `func (s *Semaphore) TryAcquire(n int64) bool`
- `func (s *Semaphore) Release(n int64) error`
- `func (s *Semaphore) Drain() int64`
- `func (s *Semaphore) Available() int64`
- `func (s *Semaphore) Capacity() int64`
- `func (s *Semaphore) Close()`

## 4. Error Handling
- `ErrTimeout`: Acquisition timed out before requested tokens became available.
- `ErrInvalidWeight`: Weight is $\le 0$ or exceeds total capacity.
- `ErrReleaseExcess`: Releasing more tokens than the semaphore's configured capacity.
- `ErrClosed`: Attempted operation on a closed semaphore or waiter unblocked due to closure.

## 5. Verification & Testing
- Unit tests verify weighted token management, non-blocking acquisition, deadline timeouts, context cancellation cleanup, FIFO ordering, capacity boundaries, and concurrency control across 100 concurrent workers.
