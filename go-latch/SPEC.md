# Technical Specification: go-latch

## 1. Domain & Purpose
`go-latch` provides robust synchronization barriers for concurrent Go applications: `CountDownLatch` and `CyclicBarrier`. It extends beyond standard `sync.WaitGroup` by providing timed waiting, context cancellation, cyclic multi-phase barrier synchronization, and broken barrier error propagation.

## 2. Architecture & Design Principles
- **CountDownLatch**:
  - One-shot synchronization latch.
  - Supports standard `Wait()`, timed `WaitTimeout(duration)`, and context-aware `WaitContext(ctx)`.
  - Atomic zero count closure eliminates deadlocks.
- **CyclicBarrier**:
  - Reusable synchronization point for $N$ parties across multi-stage iterations.
  - Executes an optional barrier action hook upon the arrival of the final party before releasing all routines.
  - Automatically advances to the next generation cycle upon party completion.
  - Broken barrier mechanics: If any party aborts or times out, the barrier breaks, releasing all other waiting parties with `ErrBrokenBarrier`.
- **Zero Heavy Dependencies**: Pure standard library (`sync`, `context`, `time`, `errors`).

## 3. API Surface
- `type CountDownLatch struct`
- `func NewCountDownLatch(count int) *CountDownLatch`
- `func (c *CountDownLatch) CountDown()`
- `func (c *CountDownLatch) GetCount() int`
- `func (c *CountDownLatch) Wait()`
- `func (c *CountDownLatch) WaitTimeout(d time.Duration) bool`
- `func (c *CountDownLatch) WaitContext(ctx context.Context) error`
- `type CyclicBarrier struct`
- `func NewCyclicBarrier(parties int, action func()) *CyclicBarrier`
- `func (b *CyclicBarrier) Await(ctx context.Context) (int, error)`
- `func (b *CyclicBarrier) Reset()`
- `func (b *CyclicBarrier) Parties() int`
- `func (b *CyclicBarrier) NumberWaiting() int`
- `func (b *CyclicBarrier) IsBroken() bool`

## 4. Error Handling
- `ErrTimeout`: Waiting on latch or barrier timed out.
- `ErrBrokenBarrier`: Another party timed out, was cancelled, or the barrier was reset while waiting.
- `ErrInvalidCount`: Initial count or party count was $\le 0$.

## 5. Verification & Testing
- Unit tests verify countdown completions, deadline timeouts, context cancellation, multi-cycle barrier iterations with atomic action counters, and broken barrier cascade releases.
