# SPEC: go-eventemitter

## 1. Metadata
- **Module**: `go-eventemitter`
- **Package**: `eventemitter`
- **Location**: `C:\Library\go-eventemitter`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `sync/atomic`, `strings`, `errors`).

---

## 2. Core Capabilities
1. **Thread-Safe Event Pub-Sub Engine**:
   - `On(event, fn)` and `Once(event, fn)` subscribe to events with automated unsubscribe closure returns.
2. **Synchronous & Asynchronous Dispatch**:
   - `Emit(event, payload)` dispatches synchronously.
   - `EmitAsync(event, payload)` executes each listener concurrently in isolated goroutines.
3. **Wildcard Routing**:
   - Supports single-prefix wildcards (`order.*` matches `order.placed`, `order.cancelled`) and catch-all `*`.
4. **Panic Isolation & Leak Warnings**:
   - Isolates listener panics to prevent process crashes.
   - Triggers `SetLeakWarningHandler` when listener registration exceeds `maxListeners` threshold.

---

## 3. Public API
```go
type Listener func(event string, payload any)

type Emitter struct
func New() *Emitter
func (e *Emitter) SetMaxListeners(n int)
func (e *Emitter) SetLeakWarningHandler(fn func(event string, count int))
func (e *Emitter) On(event string, fn Listener) (func(), error)
func (e *Emitter) Once(event string, fn Listener) (func(), error)
func (e *Emitter) Emit(event string, payload any)
func (e *Emitter) EmitAsync(event string, payload any)
func (e *Emitter) ListenerCount(event string) int
func (e *Emitter) EventNames() []string
func (e *Emitter) Clear(events ...string)
```
