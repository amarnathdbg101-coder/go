# Technical Specification: go-priorityqueue

## 1. Domain & Purpose
`go-priorityqueue` is a high-performance, thread-safe, generic binary heap Priority Queue for Go. It provides logarithmic $O(\log n)$ insertion and removal and constant $O(1)$ top-element inspection for scheduling critical tasks, event timers, and tiered order dispatching.

## 2. Architecture & Design Principles
- **Contiguous Slice Heap**: Avoids tree pointer overhead by packing binary heap nodes directly into contiguous memory slices (`left = 2i + 1`, `right = 2i + 2`, `parent = (i-1)/2`).
- **Flexible Ordering Modes**: Configurable `MaxHeap` (highest priority item popped first) or `MinHeap` (lowest priority / earliest deadline popped first).
- **Type-Safe Generics**: Holds arbitrary typed values `T` accompanied by integer priority scores without reflection.
- **Thread Safety**: Complete read/write concurrency protection via `sync.RWMutex`.
- **Zero Heavy Dependencies**: Pure Go standard library (`sync`).

## 3. API Surface
- `type Order int` (`MaxHeap`, `MinHeap`)
- `type PriorityQueue[T any] struct`
- `func New[T any]() *PriorityQueue[T]`
- `func NewWithOrder[T any](order Order) *PriorityQueue[T]`
- `func (pq *PriorityQueue[T]) Push(val T, priority int)`
- `func (pq *PriorityQueue[T]) Pop() (T, int, error)`
- `func (pq *PriorityQueue[T]) Peek() (T, int, error)`
- `func (pq *PriorityQueue[T]) Len() int`
- `func (pq *PriorityQueue[T]) IsEmpty() bool`
- `func (pq *PriorityQueue[T]) Clear()`

## 4. Error Handling
- `ErrQueueEmpty`: Invoking `Pop()` or `Peek()` on an empty queue.

## 5. Verification & Testing
- Unit tests verify MaxHeap priority ordering, MinHeap inverse ordering, Peek immutability, empty queue errors, and high-concurrency multi-worker stress invariance.
