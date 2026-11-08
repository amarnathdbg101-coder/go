# Technical Specification: go-exponentialbackoff

## 1. Domain & Purpose
`go-exponentialbackoff` is a standalone, mathematical exponential backoff and jitter calculation engine in pure Go. It mitigates the "Thundering Herd Problem" in distributed systems and microservices by applying AWS-recommended randomized delay spreads across retrying clients.

## 2. Architecture & Design Principles
- **Industry Standard Jitter Strategies**:
  - `FullJitter`: Random uniform distribution between 0 and maximum exponential ceiling.
  - `EqualJitter`: Half fixed baseline backoff plus half randomized jitter.
  - `DecorrelatedJitter`: Consecutive randomized scaling without tracking attempt counters.
  - `NoJitter`: Strict deterministic exponential curve (`base * factor^attempt`).
- **Cryptographic Randomness**: Employs CSPRNG (`crypto/rand`) for unbiased floating-point distribution.
- **Context-Aware Sleep**: Built-in `Sleep(ctx, attempt)` unblocks immediately when parent context is cancelled or times out.
- **Zero Heavy Dependencies**: Pure Go standard library (`crypto/rand`, `math`, `time`, `context`, `sync`).

## 3. API Surface
- `type Algorithm int` (`FullJitter`, `EqualJitter`, `DecorrelatedJitter`, `NoJitter`)
- `type Config struct { Base, Max time.Duration; Factor float64; Algorithm Algorithm }`
- `func DefaultConfig() Config`
- `func New(cfg Config) *Backoff`
- `func (b *Backoff) Next(attempt int) time.Duration`
- `func (b *Backoff) Sleep(ctx context.Context, attempt int) error`
- `func (b *Backoff) Reset()`

## 4. Error Handling
- `ErrContextCancelled`: Context cancelled during backoff sleep.

## 5. Verification & Testing
- Unit tests verify deterministic exponential values without jitter, mathematical bounds compliance for Full and Equal Jitter, upper boundary clamping, and context cancellation responsiveness.
