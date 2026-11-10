# SPEC: go-shadow

## 1. Scope & Purpose
`go-shadow` mirrors production HTTP traffic asynchronously to candidate backend services and identifies discrepancies in HTTP status codes, latencies, and response payloads. It enables safe Canary and Dark Launch testing in production.

It does NOT modify incoming or outgoing primary HTTP requests.

## 2. Public API Contract
```go
type DiffReport struct { ... }
type Engine struct { ... }

func New(opts ...Option) *Engine
func (e *Engine) ShouldSample() bool
func (e *Engine) Enqueue(task shadowTask)
func (e *Engine) Close() error
func Middleware(engine *Engine) func(http.Handler) http.Handler
```

## 3. Concurrency & Memory Guarantees
- Primary user requests are detached immediately; shadow dispatches occur in bounded background worker goroutines.
- If the shadow work queue exceeds capacity, tasks are dropped safely to prevent unbounded heap consumption.
- Zero external memory leaks: buffers are garbage collected after comparison.

## 4. How to Extend
- Add JSON semantic field-level diffing ignoring dynamic timestamps.
- Add OpenTelemetry trace context linking via `go-w3ctrace`.
