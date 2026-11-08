# Technical Specification: go-concurrentmap

## 1. Domain & Purpose
`go-concurrentmap` is a high-throughput, sharded concurrent hash map for Go. While `sync.Map` relies on boxed `any` types requiring runtime type assertions and suffers lock escalation when promoting entries, `go-concurrentmap` provides 32-way striped mutex partitioning with compile-time type safety via generics (`K comparable, V any`).

## 2. Architecture & Design Principles
- **Lock Striping / Sharding**: Partitions keyspace across $N$ power-of-two shards (default 32) using fast FNV-1a hashing and bitmasking (`idx = hash & (power - 1)`), allowing concurrent reads/writes on separate keys without cross-shard blocking.
- **Atomic Operations**: Safe in-shard operations including `GetOrSet`, `Pop`, and atomic callback-driven `Upsert`.
- **Early-Exit Safe Iteration**: `Iter(fn func(key K, val V) bool)` iterates safely across shards with support for short-circuiting.
- **Type-Safe Generics**: Zero interface boxing or type casting overhead.
- **Zero Heavy Dependencies**: Pure Go standard library (`hash/fnv`, `sync`, `fmt`).

## 3. API Surface
- `type ConcurrentMap[K comparable, V any] struct`
- `func New[K comparable, V any]() *ConcurrentMap[K, V]`
- `func NewWithShards[K comparable, V any](shardCount int) *ConcurrentMap[K, V]`
- `func (cm *ConcurrentMap[K, V]) Set(key K, val V)`
- `func (cm *ConcurrentMap[K, V]) Get(key K) (V, bool)`
- `func (cm *ConcurrentMap[K, V]) GetOrSet(key K, newVal V) (V, bool)`
- `func (cm *ConcurrentMap[K, V]) Has(key K) bool`
- `func (cm *ConcurrentMap[K, V]) Remove(key K) bool`
- `func (cm *ConcurrentMap[K, V]) Pop(key K) (V, bool)`
- `func (cm *ConcurrentMap[K, V]) Upsert(key K, val V, cb func(exists bool, oldVal V, newVal V) V) V`
- `func (cm *ConcurrentMap[K, V]) Count() int`
- `func (cm *ConcurrentMap[K, V]) IsEmpty() bool`
- `func (cm *ConcurrentMap[K, V]) Clear()`
- `func (cm *ConcurrentMap[K, V]) Keys() []K`
- `func (cm *ConcurrentMap[K, V]) Values() []V`
- `func (cm *ConcurrentMap[K, V]) Iter(fn func(key K, val V) bool)`

## 4. Error Handling
- `ErrKeyNotFound`: Key does not exist.

## 5. Verification & Testing
- Unit tests verify basic key/value CRUD, atomic GetOrSet, modifier-driven Upsert, safe early-exit iteration, and 20-goroutine concurrent stress testing.
