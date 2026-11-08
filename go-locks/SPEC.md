# SPEC: go-locks

## 1. Metadata
- **Module**: `go-locks`
- **Package**: `golocks`
- **Location**: `C:\Library\go-locks`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `database/sql`, `sync`, `time`, `context`, `hash/fnv`, `fmt`, `errors`).

---

## 2. Core Capabilities
1. **In-Memory `KeyedMutex` with Zero-Leak Reference Counting**:
   - Locks individual string IDs (e.g. `user_102`) without blocking distinct IDs (e.g. `user_103`).
   - Automatically deallocates mutex entries from the internal map when the reference count drops to 0, preventing memory leaks over millions of unique keys.
2. **PostgreSQL Advisory Locks (`pg_try_advisory_lock`)**:
   - True distributed locking across multiple app instances using your existing PostgreSQL database without introducing Redis or Zookeeper.
   - Deterministic 64-bit integer hashing from string keys (`KeyToInt64`) using FNV-64a.
3. **Ergonomic Execution Helpers**:
   - `WithLock(key, fn)` executes closure safely with deferred unlock.
   - `TryLock(key)` non-blocking acquisition.

---

## 3. Public API
```go
type KeyedMutex struct { ... }
func NewKeyedMutex() *KeyedMutex
func (km *KeyedMutex) Lock(key string) (unlock func())
func (km *KeyedMutex) TryLock(key string) (unlock func(), acquired bool)
func (km *KeyedMutex) WithLock(key string, fn func() error) error
func (km *KeyedMutex) ActiveKeys() int

type PostgresLocker struct { ... }
func NewPostgresLocker(lockID int64) *PostgresLocker
func NewPostgresLockerFromString(key string) *PostgresLocker
func (p *PostgresLocker) TryLock(ctx context.Context, db DBQuerier) (bool, error)
func (p *PostgresLocker) Lock(ctx context.Context, db DBQuerier, pollInterval time.Duration) error
func (p *PostgresLocker) Unlock(ctx context.Context, db DBQuerier) error
func (p *PostgresLocker) LockID() int64

func KeyToInt64(key string) int64
```
