package golocks

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"time"
)

// DBQuerier abstracts *sql.DB, *sql.Tx, and *sql.Conn.
type DBQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// KeyToInt64 computes a deterministic signed 64-bit integer from any string key using FNV-64a.
func KeyToInt64(key string) int64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return int64(hasher.Sum64())
}

// PostgresLocker manages PostgreSQL advisory locks.
type PostgresLocker struct {
	lockID int64
}

// NewPostgresLocker creates an advisory lock manager for an explicit int64 key.
func NewPostgresLocker(lockID int64) *PostgresLocker {
	return &PostgresLocker{lockID: lockID}
}

// NewPostgresLockerFromString creates an advisory lock manager hashing a string into an int64 key.
func NewPostgresLockerFromString(key string) *PostgresLocker {
	return NewPostgresLocker(KeyToInt64(key))
}

// TryLock attempts to acquire the advisory lock non-blockingly.
// Returns (true, nil) if acquired, (false, nil) if held by another process.
func (p *PostgresLocker) TryLock(ctx context.Context, db DBQuerier) (bool, error) {
	var acquired bool
	err := db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", p.lockID).Scan(&acquired)
	if err != nil {
		return false, fmt.Errorf("locks: pg_try_advisory_lock query failed: %w", err)
	}
	return acquired, nil
}

// Lock blocks until the advisory lock is acquired or ctx is cancelled.
func (p *PostgresLocker) Lock(ctx context.Context, db DBQuerier, pollInterval time.Duration) error {
	if pollInterval <= 0 {
		pollInterval = 50 * time.Millisecond
	}

	for {
		acquired, err := p.TryLock(ctx, db)
		if err != nil {
			return err
		}
		if acquired {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// Unlock releases the advisory lock in PostgreSQL.
func (p *PostgresLocker) Unlock(ctx context.Context, db DBQuerier) error {
	var released bool
	err := db.QueryRowContext(ctx, "SELECT pg_advisory_unlock($1)", p.lockID).Scan(&released)
	if err != nil {
		return fmt.Errorf("locks: pg_advisory_unlock query failed: %w", err)
	}
	if !released {
		return ErrLockNotHeld
	}
	return nil
}

// LockID returns the internal 64-bit advisory lock ID.
func (p *PostgresLocker) LockID() int64 {
	return p.lockID
}
