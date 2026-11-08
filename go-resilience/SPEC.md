# Technical Specification: go-resilience

## 1. Domain & Purpose
`go-resilience` is a unified, composable fault-tolerance pipeline for Go microservices. It brings together Concurrency Bulkheading, Circuit Breaking, Exponential Backoff Retries with Full Jitter, Execution Timeouts, and Graceful Fallback Degradation into an ergonomic, type-safe execution engine (`Policy[T]`).

## 2. Architecture & Design Principles
- **Fluent Onion-Layer Composition**: Coordinates incoming execution through layered stages:
  1. *Bulkhead Barrier*: Throttles concurrent executions to protect local memory and downstream capacity (`ErrBulkheadFull`).
  2. *Circuit Breaker*: Fails fast on degraded dependencies (`ErrCircuitOpen`), automatically transitioning through Closed, Open, and Half-Open canary states.
  3. *Retry with Full Jitter*: Retries transient failures with randomized exponential backoff to eliminate thundering herds.
  4. *Timeout Deadline*: Wraps each execution attempt in a bounded `context.WithTimeout`.
  5. *Graceful Fallback*: Diverts failed executions to user-provided fallback handlers without leaking errors to end users.
- **Generic Type System**: Parameterized by return type `T` with full compile-time safety.
- **Zero Heavy Dependencies**: Pure standard library (`context`, `sync`, `time`, `crypto/rand`, `math/big`).

## 3. API Surface
- `type RetryConfig struct`
- `type CircuitBreakerConfig struct`
- `type Policy[T any] struct`
- `func NewPolicy[T any]() *Policy[T]`
- `func (p *Policy[T]) WithTimeout(d time.Duration) *Policy[T]`
- `func (p *Policy[T]) WithRetry(cfg RetryConfig) *Policy[T]`
- `func (p *Policy[T]) WithCircuitBreaker(cfg CircuitBreakerConfig) *Policy[T]`
- `func (p *Policy[T]) WithBulkhead(maxConcurrent int) *Policy[T]`
- `func (p *Policy[T]) WithFallback(fn func(ctx context.Context, err error) (T, error)) *Policy[T]`
- `func (p *Policy[T]) Execute(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)`

## 4. Error Handling
- `ErrBulkheadFull`: Maximum concurrent execution quota reached.
- `ErrCircuitOpen`: Circuit breaker is open and fast-rejecting execution.
- `ErrTimeout`: Execution attempt exceeded deadline duration.
- `ErrMaxAttemptsReached`: All retry attempts were exhausted.

## 5. Verification & Testing
- Unit tests verify happy path completion, transient retry recovery, per-attempt deadline enforcement, bulkhead concurrency rejection, circuit breaker trip transitions, and fallback handler substitution.
