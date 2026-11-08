# SPEC: go-ratelimit

## 1. Scope & Purpose
High-performance, zero-memory-leak rate limiting library for Go microservices and APIs.
- In Scope: Token Bucket & Sliding Window algorithms, Client IP / Header / Route key strategies, RFC 6585 header injection, In-memory RAM store + Distributed Redis Lua script store, net/http middleware.
- Out of Scope: Hardware firewall rule creation.

## 2. Public API Contract
- `New(store Store, rule Rule, opts ...Option) (*Limiter, error)`
- `NewIPLimiter(capacity int, window time.Duration) (*Limiter, error)`
- `Limiter.Middleware(next http.Handler) http.Handler`
- `Store` Interface: `Take(ctx, key, rule) (Result, error)`, `Close() error`
- `MemoryStore`: High-speed thread-safe token bucket store with automatic ticker cleanup
- `RedisStore`: Distributed rate limit store using atomic Redis Lua Script

## 3. Concurrency & Memory Guarantees
- Thread-safe storage with `sync.Mutex` token bucket calculations.
- Evaluation speed: < 210 nanoseconds per check.
- MemoryStore runs a 3-minute cleanup ticker goroutine to evict inactive buckets older than 10 minutes.
- Graceful shutdown via `Limiter.Close()`.

## 4. How to Extend
- Implement `Store` interface for new storage engines (Memcached, DynamoDB).
- Implement custom `KeyFunc` for proprietary authentication headers or JWT claims.
