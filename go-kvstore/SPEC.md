# Technical Specification: go-kvstore

## 1. Domain & Purpose
`go-kvstore` is a high-concurrency, sharded in-memory key-value cache engine for Go. It provides sub-microsecond lookups, dual TTL expiration (lazy check on access + periodic active background janitor), lock-striped sharding to eradicate mutex contention across high-core architectures, and atomic synchronization primitives (`GetOrSet`, `CompareAndSwap`, `Increment`).

## 2. Architecture & Design Principles
- **Lock Striping / Sharding**: Divides keyspace across $N$ power-of-two shards (default 32) using fast FNV-1a hashing and bitwise masking, allowing concurrent read/write operations on independent keys without blocking each other.
- **Dual Expiration Mechanics**: Expired items are purged lazily during `Get` requests and actively via a background ticker sweeping across shards.
- **Atomic Operations**: Thread-safe implementations of `CompareAndSwap`, `GetOrSet`, and multi-goroutine integer `Increment` without race conditions.
- **Zero Heavy Dependencies**: Pure Go standard library (`sync`, `time`, `hash/fnv`).

## 3. API Surface
- `type Config struct { ShardCount int; CleanupInterval time.Duration }`
- `func DefaultConfig() Config`
- `func New(cfg Config) *Store`
- `func (s *Store) Set(key string, val any, ttl time.Duration)`
- `func (s *Store) Get(key string) (any, bool)`
- `func (s *Store) GetOrSet(key string, val any, ttl time.Duration) (actual any, loaded bool)`
- `func (s *Store) CompareAndSwap(key string, oldVal, newVal any) bool`
- `func (s *Store) Increment(key string, delta int64) (int64, error)`
- `func (s *Store) Delete(key string) bool`
- `func (s *Store) Len() int`
- `func (s *Store) Flush()`
- `func (s *Store) Close()`

## 4. Error Handling
- `ErrKeyNotFound`: Key does not exist or has expired.
- `ErrNotAnInteger`: Increment invoked on non-integer value.
- `ErrStoreClosed`: Operations attempted on closed store.

## 5. Verification & Testing
- Unit tests verify basic storage/retrieval, lazy & active TTL expiration, CAS state transitions, atomic counter increments, and 20-worker concurrent load stress testing.
