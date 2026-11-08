# Technical Specification: go-ringbuffer

## 1. Domain & Purpose
`go-ringbuffer` provides a thread-safe, generic circular FIFO ring buffer in pure Go. It delivers pre-allocated memory efficiency for streaming data pipelines, telemetry sliding windows, audio/video frame buffering, and fixed-capacity log tails without heap reallocations.

## 2. Architecture & Design Principles
- **Dual Saturation Modes**: Supports both bounded rejection (`ErrBufferFull`) and overwrite mode (evicts oldest items on buffer saturation).
- **Zero Heap Reallocation**: Pre-allocates slice memory upon initialization; internal head and tail pointers wrap around cyclically.
- **Garbage Collection Friendly**: Clears array references on `Pop` to ensure unreferenced objects are reclaimed by Go runtime GC.
- **Type-Safe Generics**: Works with any payload `T` without type casting.
- **Zero Heavy Dependencies**: Pure Go standard library (`sync`).

## 3. API Surface
- `type RingBuffer[T any] struct`
- `func New[T any](capacity int, overwrite bool) (*RingBuffer[T], error)`
- `func (r *RingBuffer[T]) Push(item T) error`
- `func (r *RingBuffer[T]) Pop() (T, error)`
- `func (r *RingBuffer[T]) Peek() (T, error)`
- `func (r *RingBuffer[T]) Slice() []T`
- `func (r *RingBuffer[T]) Len() int`
- `func (r *RingBuffer[T]) Cap() int`
- `func (r *RingBuffer[T]) IsFull() bool`
- `func (r *RingBuffer[T]) IsEmpty() bool`
- `func (r *RingBuffer[T]) Reset()`

## 4. Error Handling
- `ErrBufferFull`: Push attempted on full buffer with overwrite disabled.
- `ErrBufferEmpty`: Pop attempted on empty buffer.
- `ErrInvalidCapacity`: Capacity <= 0.

## 5. Verification & Testing
- Unit tests verify FIFO ordering, Peek preservation, bounded rejection, overwrite FIFO sequence, and multi-worker concurrent read/write stress.
