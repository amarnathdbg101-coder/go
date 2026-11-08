# SPEC: go-orderedmap

## 1. Metadata
- **Module**: `go-orderedmap`
- **Package**: `orderedmap`
- **Location**: `C:\Library\go-orderedmap`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `encoding/json`, `bytes`, `fmt`, `sort`).

---

## 2. Core Capabilities
1. **O(1) Insertion-Preserving Hash Map**:
   - Implements doubly-linked list nodes coupled with a generic Go hash map to achieve $O(1)$ lookups, sets, and deletions while preserving strictly deterministic insertion order.
2. **First-Class JSON Serialization**:
   - Custom `MarshalJSON` guarantees keys are written to JSON in exact insertion order.
   - Custom `UnmarshalJSON` parses arbitrary JSON objects directly into an `OrderedMap[string, V]` without losing property sequence.
3. **Thread-Safe & Lock-Free Modes**:
   - `New[K, V]()` provides zero-locking high-throughput single-threaded performance.
   - `NewConcurrent[K, V]()` wraps operations in `sync.RWMutex` for concurrent multi-goroutine access.
4. **In-Place Sorting & Traversal**:
   - `Sort(lessFunc)` allows reordering entries in-place.
   - `ForEach(fn)` supports early-exit traversal.

---

## 3. Public API
```go
type Element[K comparable, V any] struct {
    Key   K
    Value V
    func (e *Element[K, V]) Next() *Element[K, V]
    func (e *Element[K, V]) Prev() *Element[K, V]
}

type Entry[K comparable, V any] struct {
    Key   K `json:"key"`
    Value V `json:"value"`
}

type Map[K comparable, V any] struct
func New[K comparable, V any]() *Map[K, V]
func NewConcurrent[K comparable, V any]() *Map[K, V]
func (m *Map[K, V]) Set(key K, value V)
func (m *Map[K, V]) Get(key K) (V, bool)
func (m *Map[K, V]) GetOrDefault(key K, defaultVal V) V
func (m *Map[K, V]) Has(key K) bool
func (m *Map[K, V]) Delete(key K) bool
func (m *Map[K, V]) Len() int
func (m *Map[K, V]) Clear()
func (m *Map[K, V]) Front() *Element[K, V]
func (m *Map[K, V]) Back() *Element[K, V]
func (m *Map[K, V]) Keys() []K
func (m *Map[K, V]) Values() []V
func (m *Map[K, V]) Entries() []Entry[K, V]
func (m *Map[K, V]) ForEach(fn func(key K, value V) bool)
func (m *Map[K, V]) Sort(less func(a, b Entry[K, V]) bool)
func (m *Map[K, V]) MarshalJSON() ([]byte, error)
func (m *Map[string, V]) UnmarshalJSON(data []byte) error
```
