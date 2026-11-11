# SPEC: go-cachekit/data

## 1. Scope & Purpose
The `data` package is a small facade for the multi-tier cache engine in the consolidated `go-cachekit` module.
It exposes the engine and store types through `github.com/amarnathdbg101-coder/go/go-cachekit/data/cache`.

## 2. Public API Sub-packages
- `github.com/amarnathdbg101-coder/go/go-cachekit/data`: Cache engine and store facade
- `github.com/amarnathdbg101-coder/go/go-cachekit/data/cache`: L1 RAM + L2 Redis, Singleflight Dogpile protection, Generics `GetOrSet[T]`
- `github.com/amarnathdbg101-coder/go/go-cachekit/lru`: LRU cache policy
- `github.com/amarnathdbg101-coder/go/go-cachekit/lfu`: LFU cache policy
- `github.com/amarnathdbg101-coder/go/go-cachekit/tinylfu`: TinyLFU cache policy

## 3. Concurrency & Memory Guarantees
- Sub-microsecond evaluation speed.
- Singleflight coalesces 1,000 concurrent DB requests into 1 query.
- Thread-safe storage with sync.RWMutex.
