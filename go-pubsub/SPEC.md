# SPEC: go-pubsub

## 1. Metadata
- **Module**: `go-pubsub`
- **Package**: `gopubsub`
- **Location**: `C:\Library\go-pubsub`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `sync/atomic`).

---

## 2. Core Capabilities
1. **Generic Type-Safe PubSub**:
   - `PubSub[T any]` provides broadcast message fanout to all active subscribers of a topic.
2. **Non-Blocking Slow-Consumer Mitigation**:
   - `DropOldest`: drops stale messages to preserve fresh telemetry/GPS points without blocking publishers or other subscribers.
   - `DropNew`: drops new messages if the subscriber queue is saturated.
3. **Channel-Based Subscriptions**:
   - Clean `<-chan T` ergonomics for WebSocket connection loops and goroutine workers.
4. **Leak-Proof Unsubscribe**:
   - Safely closes channel and removes subscriber tracking upon `sub.Unsubscribe()`.

---

## 3. Public API
```go
type OverflowStrategy int
const (
    DropOldest OverflowStrategy = iota
    DropNew
)

type Subscription[T any] struct {
    Ch <-chan T
}
func (s *Subscription[T]) Unsubscribe()

type PubSub[T any] struct { ... }
func New[T any](bufferSize int, strategy OverflowStrategy) *PubSub[T]
func (ps *PubSub[T]) Subscribe(topic string) *Subscription[T]
func (ps *PubSub[T]) Publish(topic string, msg T) int
func (ps *PubSub[T]) SubscribersCount(topic string) int
func (ps *PubSub[T]) Close()
```
