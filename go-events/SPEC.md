# SPEC: go-events

## 1. Metadata
- **Module**: `go-events`
- **Package**: `goevents`
- **Location**: `C:\Library\go-events`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `context`, `strings`, `fmt`, `errors`).

---

## 2. Core Capabilities
1. **Hierarchical Topic Pattern Matching**:
   - Single segment wildcard `*` (e.g. `order.*` matches `order.created`, `order.paid`).
   - Multi-segment wildcard `**` or `>` (e.g. `order.**` matches `order.item.added`).
2. **Synchronous & Asynchronous Dispatching**:
   - `Publish`: blocking synchronous dispatch to all matching listeners.
   - `PublishAsync`: non-blocking enqueue into a bounded worker pool channel.
3. **Panic Isolation**:
   - Individual listener panics are isolated via `recover()` without interrupting other listeners or crashing the server.
4. **Dead-Letter Queue (DLQ)**:
   - Automated routing of unhandled events or repeated handler errors to DLQ callback.
5. **Middleware Pipeline**:
   - Interceptors for distributed tracing, correlation IDs, and execution latency.

---

## 3. Public API
```go
type Event struct {
    Topic     string            `json:"topic"`
    Payload   any               `json:"payload"`
    Timestamp time.Time         `json:"timestamp"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

func NewEvent(topic string, payload any) Event
func (e Event) WithMetadata(key, value string) Event

type HandlerFunc func(ctx context.Context, e Event) error
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type Config struct {
    WorkerCount int
    QueueSize   int
    OnPanic     func(topic string, recovered any)
    OnDLQ       func(event Event, err error)
}

type Bus struct { ... }
func New(cfg ...Config) *Bus
func (b *Bus) Subscribe(pattern string, handler HandlerFunc) (*subscription, error)
func (b *Bus) Publish(ctx context.Context, event Event) error
func (b *Bus) PublishAsync(event Event) error
func (b *Bus) Use(middlewares ...MiddlewareFunc)
func (b *Bus) Close() error
```
