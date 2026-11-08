# Technical Specification: go-slicekit

## 1. Domain & Purpose
`go-slicekit` is a high-performance, zero-allocation generic slice manipulation library for Go 1.22+. It provides missing collection utilities (Chunk, Flatten, Unique, UniqueBy, GroupBy, Partition, Reverse, Compact, Shuffle, Map, Filter, Reduce) with compile-time type safety.

## 2. Architecture & Design Principles
- **Generic Type System**: Full support for any types `T` and comparable types `T comparable`.
- **Pre-Allocated Memory Sizing**: Functions like `Chunk`, `Flatten`, and `Unique` pre-calculate capacity whenever possible to avoid multiple reallocation copies.
- **Order Preservation**: `Unique` and `UniqueBy` strictly maintain the first-seen insertion order.
- **Crypto-Secure Shuffle**: Uses `crypto/rand` Fisher-Yates algorithm for unbiased slice shuffling without math/rand global seed locks.
- **Pure Standard Library**: Zero external dependencies (`crypto/rand`, `math/big`).

## 3. API Surface
- `func Chunk[T any](slice []T, size int) [][]T`
- `func Flatten[T any](slices [][]T) []T`
- `func Unique[T comparable](slice []T) []T`
- `func UniqueBy[T any, K comparable](slice []T, keyFn func(T) K) []T`
- `func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T`
- `func Partition[T any](slice []T, predicate func(T) bool) ([]T, []T)`
- `func Reverse[T any](slice []T) []T`
- `func ReverseInPlace[T any](slice []T)`
- `func Compact[T comparable](slice []T) []T`
- `func Filter[T any](slice []T, predicate func(T) bool) []T`
- `func Map[T any, R any](slice []T, transform func(T) R) []R`
- `func Reduce[T any, R any](slice []T, initial R, reducer func(R, T) R) R`
- `func Find[T any](slice []T, predicate func(T) bool) (T, bool)`
- `func FindIndex[T any](slice []T, predicate func(T) bool) int`
- `func ContainsBy[T any](slice []T, predicate func(T) bool) bool`
- `func Shuffle[T any](slice []T)`

## 4. Error Handling
- `ErrInvalidChunkSize`: Chunk size parameter is $\le 0$.
- `ErrEmptySlice`: Slice is empty when an element is strictly required.

## 5. Verification & Testing
- Unit tests verify batch chunking, 2D array flattening, unique deduplication, struct grouping, partition splits, reversals, zero-value compaction, transformations, and Fisher-Yates shuffling.
