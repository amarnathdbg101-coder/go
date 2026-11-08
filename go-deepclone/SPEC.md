# Technical Specification: go-deepclone

## 1. Domain & Purpose
`go-deepclone` is a reflection-safe deep cloning engine for Go. It creates true independent deep copies of complex data structures (pointers, slices, maps, nested structs, and interfaces) preventing unexpected state mutations, data races, and shallow copy bugs, while safely detecting and preserving circular references.

## 2. Architecture & Design Principles
- **Recursive Reflection Duplication**: Allocates completely distinct memory blocks for all pointers, maps, and slice buffers.
- **Cycle-Detection Pointer Tracking**: Maintains a `visited map[uintptr]reflect.Value` to detect pointer loops (e.g. doubly-linked lists or cyclic graphs), preventing stack overflow and recreating the exact cycle topology.
- **Nil Safety**: Accurately preserves nil pointers, nil slices, and nil maps without allocating empty headers.
- **Type-Safe Generics**: `Clone[T](src T) T` returns the exact compile-time type without runtime type assertion boilerplate.
- **Zero Heavy Dependencies**: Pure standard library (`reflect`).

## 3. API Surface
- `func Clone[T any](src T) T`
- `func CloneValue(src any) any`

## 4. Error Handling
- Safe traversal with nil propagation and zero-value fallback on invalid reflect values.

## 5. Verification & Testing
- Unit tests verify deep memory independence of nested structs, slice buffers, map entries, nil references, and recursive circular pointer topologies.
