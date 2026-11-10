# SPEC: go-crdt

## 1. Scope & Purpose
`go-crdt` provides conflict-free replicated data structures for distributed and offline-first Go applications. It guarantees strong eventual consistency across mobile clients and distributed microservices without distributed locks or two-phase commit protocols.

It does NOT provide network transport or database persistence directly.

## 2. Public API Contract
```go
type PNCounter struct { ... }
func NewPNCounter(nodeID string) *PNCounter
func (c *PNCounter) Inc(delta uint64)
func (c *PNCounter) Dec(delta uint64)
func (c *PNCounter) Value() int64
func (c *PNCounter) Merge(other *PNCounter) error

type LWWRegister[T any] struct { ... }
func NewLWWRegister[T any](nodeID string, initialValue T) *LWWRegister[T]
func (r *LWWRegister[T]) Set(val T)
func (r *LWWRegister[T]) Value() T
func (r *LWWRegister[T]) Merge(other *LWWRegister[T]) error

type ORSet[T comparable] struct { ... }
func NewORSet[T comparable](nodeID string) *ORSet[T]
func (s *ORSet[T]) Add(elem T)
func (s *ORSet[T]) Remove(elem T)
func (s *ORSet[T]) Contains(elem T) bool
func (s *ORSet[T]) Elements() []T
func (s *ORSet[T]) Merge(other *ORSet[T]) error
```

## 3. Concurrency & Memory Guarantees
- Thread-safe mutation via `sync.RWMutex`.
- All merge operations obey join-semilattice algebraic properties.
- State is bounded by the number of active replica nodes and unique live elements.

## 4. How to Extend
- Implement Delta-State CRDTs for low-bandwidth cellular transport.
- Implement JSON CRDT Tree for rich collaborative document editing.
