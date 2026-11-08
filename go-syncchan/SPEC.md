# SPEC: go-syncchan

## 1. Metadata
- **Module**: `go-syncchan`
- **Package**: `syncchan`
- **Location**: `C:\Library\go-syncchan`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `context`, `sync`).

---

## 2. Core Capabilities
1. **Goroutine-Leak-Proof Channel Combinators**:
   - Every combinator multiplexes internal reading with `ctx.Done()`, ensuring zero zombie goroutines when parent contexts expire.
2. **Generic Stream Operations**:
   - `FanIn(ctx, chans...)`: Merges $N$ input streams of type `T` into a consolidated output channel.
   - `FanOut(ctx, in, n)`: Broadcasts values to $N$ independent subscriber channels.
   - `Tee(ctx, in)`: Splits an input stream into two distinct mirrors.
   - `Take(ctx, in, n)` & `Drop(ctx, in, n)`: Bounded window extraction.
   - `OrDone(ctx, in)`: Non-blocking context-aware read guard.
   - `Buffer(ctx, in, size)`: Dynamic backpressure buffer.

---

## 3. Public API
```go
func OrDone[T any](ctx context.Context, in <-chan T) <-chan T
func FanIn[T any](ctx context.Context, chans ...<-chan T) <-chan T
func Tee[T any](ctx context.Context, in <-chan T) (<-chan T, <-chan T)
func FanOut[T any](ctx context.Context, in <-chan T, n int) []<-chan T
func Take[T any](ctx context.Context, in <-chan T, n int) <-chan T
func Drop[T any](ctx context.Context, in <-chan T, n int) <-chan T
func Buffer[T any](ctx context.Context, in <-chan T, size int) <-chan T
func Collect[T any](in <-chan T) []T
```
