# Technical Specification: go-timingwheel

## 1. Domain & Purpose
`go-timingwheel` is a high-capacity, $O(1)$ Hashed Timing Wheel timer scheduler for Go. Designed for microsecond/millisecond network timeouts, WebSocket heartbeat monitors, and delayed task scheduling, it provides constant-time $O(1)$ timer insertion and cancellation, avoiding the Go runtime timer heap lock contention under hundreds of thousands of concurrent timers.

## 2. Architecture & Design Principles
- **Hashed Circular Slots**: Bucketed doubly-linked lists representing discrete time tick intervals.
- **Multi-Round Expiration**: Calculates remaining cycle rotations (`rounds`) to support arbitrarily large expiration durations beyond a single wheel revolution.
- **Constant Time $O(1)$ Insertion & Cancellation**: Starting and stopping a timer operates in $O(1)$ by unlinking from its doubly-linked slot list.
- **Decoupled Asynchronous Execution**: Timer callbacks execute in separate goroutines outside the internal mutex lock, preventing slow task blocking.
- **Zero Heavy Dependencies**: Pure standard library (`sync`, `time`).

## 3. API Surface
- `type Timer struct`
- `func (t *Timer) Stop() bool`
- `type TimingWheel struct`
- `func New(tick time.Duration, wheelSize int) *TimingWheel`
- `func (tw *TimingWheel) Start()`
- `func (tw *TimingWheel) Stop()`
- `func (tw *TimingWheel) AfterFunc(d time.Duration, fn func()) (*Timer, error)`

## 4. Error Handling
- `ErrInvalidTick`: Duration or wheel size is $\le 0$.
- `ErrWheelStopped`: Scheduling attempted on a terminated timing wheel.

## 5. Verification & Testing
- Unit tests verify delay accuracy, cancellation before execution, multi-round schedule cycles, and concurrent task scheduling under 100 parallel workers.
