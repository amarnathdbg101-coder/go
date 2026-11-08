# SPEC: go-cachekit

## 1. Scope & Purpose
Ultra-fast, generics-powered multi-tier caching library for Go microservices.
- In Scope: Read-through caching, Singleflight cache stampede / thundering herd shield, Stale-While-Revalidate (SWR) async background revalidation, In-memory RAM store + Redis L2 store, TwoTierStore hybrid backfilling.
- Out of Scope: Database ORM / SQL query parsing.

## 2. Public API Contract
- `New(store Store) (*Engine, error)`
- `Get[T any](ctx, engine, key) (T, error)`
- `Set[T any](ctx, engine, key, value, ttl) error`
- `GetOrSet[T any](ctx, engine, key, ttl, fetcher) (T, error)`
- `GetOrSetSWR[T any](ctx, engine, key, freshTTL, staleTTL, fetcher) (T, error)`
- `Typed[T any](engine) *TypedCache[T]`
- `Store` Interface: `Get`, `Set`, `Delete`, `DeletePrefix`, `Clear`, `Close`
- `MemoryStore`: High-speed thread-safe RAM cache with passive + active background eviction
- `RedisStore`: Production L2 remote cache backed by Redis
- `TwoTierStore`: L1 (RAM) + L2 (Redis) hybrid store with automatic L1 backfill

## 3. Concurrency & Memory Guarantees
- Sub-microsecond RAM read latency (~53.5 nanoseconds per hit).
- Singleflight coalesces 1,000 concurrent requests into 1 database query.
- MemoryStore runs background cleanup ticker goroutine to evict expired keys.
- Graceful teardown via `Close()`.

## 4. How to Extend
- Implement `Store` interface for new cache drivers (Memcached, BadgerDB).
