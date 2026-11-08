# Technical Specification: go-debounce

## 1. Domain & Purpose
`go-debounce` provides thread-safe Debounce and Throttle execution controllers in pure Go. It protects backend databases, message brokers, autosave pipelines, and third-party APIs from overload during rapid bursts of incoming events by regulating invocation frequencies.

## 2. Architecture & Design Principles
- **Debouncer**: Delays invocation of an action until a period of silence has elapsed. Successive calls within the wait duration reset the countdown timer.
- **Throttler**: Enforces a strict rate ceiling, guaranteeing an action runs at most once per time window without queuing delays on the first call.
- **Keyed Debouncing**: Isolates debouncing timers across independent keys (e.g. per user ID or document ID), preventing global lockouts across multi-tenant workloads.
- **Zero Heavy Dependencies**: Pure Go standard library (`time`, `sync`).

## 3. API Surface
- `type Debouncer struct`
- `func New(wait time.Duration) *Debouncer`
- `func (d *Debouncer) Trigger(action func())`
- `func (d *Debouncer) Cancel()`
- `type Throttler struct`
- `func NewThrottler(wait time.Duration) *Throttler`
- `func (t *Throttler) Trigger(action func()) bool`
- `type KeyedDebouncer struct`
- `func NewKeyed(wait time.Duration) *KeyedDebouncer`
- `func (kd *KeyedDebouncer) Trigger(key string, action func())`
- `func (kd *KeyedDebouncer) Cancel(key string)`
- `func (kd *KeyedDebouncer) CancelAll()`

## 4. Error Handling
- `ErrInvalidDuration`: Duration <= 0.
- `ErrActionNil`: Action callback is nil.

## 5. Verification & Testing
- Unit tests verify burst coalescing (10 rapid calls -> 1 execution), cancellation cleanup, throttle time windows, and multi-key isolation.
