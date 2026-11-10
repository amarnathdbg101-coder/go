# SPEC: go-distlock

## 1. Scope & Purpose
`go-distlock` provides distributed mutual exclusion with fencing token safety guarantees for Go microservices. It targets single-leader election, batch processing deduplication, and resource lease management.

It does NOT manage consensus clusters (use Raft/Paxos for distributed log consensus).

## 2. Public API Contract
```go
type Driver interface { ... }
type Lock struct { ... }
type Locker struct { ... }

func Acquire(ctx context.Context, key string, opts ...Option) (*Lock, error)
func (l *Lock) FencingToken() uint64
func (l *Lock) Key() string
func (l *Lock) Owner() string
func (l *Lock) Release() error
```

## 3. Concurrency & Memory Guarantees
- Driver storage operations are synchronized via `sync.Mutex`.
- Fencing token counter strictly increments monotonically.
- Heartbeat renewal loop terminates deterministically when `Release()` is called.

## 4. How to Extend
- Implement Redis driver using Redlock or single-instance Redis Lua script with `SET NX PX` and `PEXPIRE`.
- Implement PostgreSQL advisory locks driver using `pg_try_advisory_lock`.
