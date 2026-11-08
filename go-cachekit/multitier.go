package cachekit

import (
	"context"
	"time"
)

// TwoTierStore combines an ultra-fast L1 In-Memory cache with an L2 Distributed/Remote cache (e.g. Redis).
// Reads: L1 (RAM ~2µs) -> if miss -> L2 (Remote ~2ms) -> backfills L1 -> return.
// Writes: Writes to both L1 and L2 simultaneously.
// Deletes: Evicts from both L1 and L2 to prevent cache inconsistency.
type TwoTierStore struct {
	l1 Store // In-Memory Cache
	l2 Store // Remote / Redis Cache
}

// NewTwoTierStore initializes a two-tier hybrid caching store.
func NewTwoTierStore(l1 Store, l2 Store) *TwoTierStore {
	return &TwoTierStore{
		l1: l1,
		l2: l2,
	}
}

// Get reads from L1 first; on miss, falls back to L2 and automatically backfills L1.
func (s *TwoTierStore) Get(ctx context.Context, key string) ([]byte, error) {
	// 1. Check L1 In-Memory
	if s.l1 != nil {
		val, err := s.l1.Get(ctx, key)
		if err == nil {
			return val, nil
		}
	}

	// 2. Check L2 Remote/Distributed
	if s.l2 != nil {
		val, err := s.l2.Get(ctx, key)
		if err == nil {
			// Auto backfill L1 so subsequent requests are served from RAM
			if s.l1 != nil {
				// Default backfill TTL (e.g., 5 minutes or L2 TTL)
				_ = s.l1.Set(ctx, key, val, 5*time.Minute)
			}
			return val, nil
		}
	}

	return nil, ErrCacheMiss
}

// Set saves data to both L1 and L2 stores.
func (s *TwoTierStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	var err1, err2 error

	if s.l1 != nil {
		err1 = s.l1.Set(ctx, key, value, ttl)
	}
	if s.l2 != nil {
		err2 = s.l2.Set(ctx, key, value, ttl)
	}

	if err1 != nil {
		return err1
	}
	return err2
}

// Delete evicts key from both L1 and L2.
func (s *TwoTierStore) Delete(ctx context.Context, key string) error {
	if s.l1 != nil {
		_ = s.l1.Delete(ctx, key)
	}
	if s.l2 != nil {
		_ = s.l2.Delete(ctx, key)
	}
	return nil
}

// DeletePrefix evicts all matching keys from both L1 and L2.
func (s *TwoTierStore) DeletePrefix(ctx context.Context, prefix string) error {
	if s.l1 != nil {
		_ = s.l1.DeletePrefix(ctx, prefix)
	}
	if s.l2 != nil {
		_ = s.l2.DeletePrefix(ctx, prefix)
	}
	return nil
}

// Clear flushes both tiers.
func (s *TwoTierStore) Clear(ctx context.Context) error {
	if s.l1 != nil {
		_ = s.l1.Clear(ctx)
	}
	if s.l2 != nil {
		_ = s.l2.Clear(ctx)
	}
	return nil
}

// Close closes both tiers.
func (s *TwoTierStore) Close() error {
	if s.l1 != nil {
		_ = s.l1.Close()
	}
	if s.l2 != nil {
		_ = s.l2.Close()
	}
	return nil
}
