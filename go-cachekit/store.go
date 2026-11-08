package cachekit

import (
	"context"
	"time"
)

// Store defines the low-level byte-level persistence contract for any caching backend.
// In-Memory RAM, Redis, Memcached, ya multi-tier setups ke liye ye interface implement hota hai.
type Store interface {
	// Get retrieves cached raw bytes by key. Returns ErrCacheMiss if not found or expired.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores raw bytes with a time-to-live expiration.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key from cache. No error is returned if key does not exist.
	Delete(ctx context.Context, key string) error

	// DeletePrefix removes all keys matching the specified prefix (e.g. "products:").
	DeletePrefix(ctx context.Context, prefix string) error

	// Clear flushes all cached entries from this store.
	Clear(ctx context.Context) error

	// Close terminates any background workers and frees resources.
	Close() error
}
