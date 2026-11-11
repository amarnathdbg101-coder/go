package authkit

import (
	"context"
	"sync"
	"time"
)

// RevocationStore defines the contract for persisting revoked token IDs (JTIs)
// to support single-use Refresh Token Rotation and secure Logout.
type RevocationStore interface {
	// Revoke marks a specific token identifier (JTI) as revoked until its natural expiry time.
	Revoke(ctx context.Context, tokenID string, expiresAt time.Time) error

	// IsRevoked checks if a token ID has been revoked.
	IsRevoked(ctx context.Context, tokenID string) (bool, error)

	// Close terminates any background maintenance or cleanup routines.
	Close() error
}

// MemoryRevocationStore is a high-performance, thread-safe, in-memory implementation of RevocationStore.
// An internal ticker periodically evicts expired entries to prevent memory leaks.
type MemoryRevocationStore struct {
	mu       sync.RWMutex
	revoked  map[string]time.Time
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewMemoryRevocationStore creates a new MemoryRevocationStore and starts the background eviction routine.
func NewMemoryRevocationStore() *MemoryRevocationStore {
	store := &MemoryRevocationStore{
		revoked:  make(map[string]time.Time),
		stopChan: make(chan struct{}),
	}
	go store.startEvictionLoop()
	return store
}

// Revoke inserts the token ID into the revocation map.
func (s *MemoryRevocationStore) Revoke(ctx context.Context, tokenID string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[tokenID] = expiresAt
	return nil
}

// IsRevoked returns true if the token ID is explicitly revoked and has not yet passed expiry.
func (s *MemoryRevocationStore) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, found := s.revoked[tokenID]
	if !found {
		return false, nil
	}
	if time.Now().After(exp) {
		return false, nil
	}
	return true, nil
}

// startEvictionLoop periodically sweeps and deletes expired token IDs from RAM.
func (s *MemoryRevocationStore) startEvictionLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.evictExpired()
		case <-s.stopChan:
			return
		}
	}
}

// evictExpired purges past-due tokens.
func (s *MemoryRevocationStore) evictExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, exp := range s.revoked {
		if now.After(exp) {
			delete(s.revoked, id)
		}
	}
}

// Close gracefully stops the background eviction goroutine.
func (s *MemoryRevocationStore) Close() error {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
	return nil
}
