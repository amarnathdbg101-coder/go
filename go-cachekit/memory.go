package cachekit

import (
	"context"
	"strings"
	"sync"
	"time"
)

// memoryItem holds cached bytes along with its exact expiration timestamp.
type memoryItem struct {
	value     []byte
	expiresAt time.Time
}

// MemoryStore is a thread-safe, high-speed in-memory cache store.
// Pure Go Standard library (sync.RWMutex) se bana hai, zero external dependencies ke saath.
type MemoryStore struct {
	mu       sync.RWMutex
	items    map[string]memoryItem
	stopChan chan struct{}
	stopOnce sync.Once
}

// MemoryOptions configures the MemoryStore behavior.
type MemoryOptions struct {
	// CleanupInterval controls how often background goroutine evicts expired entries. Default: 1 minute.
	CleanupInterval time.Duration
}

// NewMemoryStore initializes an in-memory cache and starts an automatic background cleanup worker.
func NewMemoryStore(opts ...MemoryOptions) *MemoryStore {
	cleanupInterval := 1 * time.Minute
	if len(opts) > 0 && opts[0].CleanupInterval > 0 {
		cleanupInterval = opts[0].CleanupInterval
	}

	s := &MemoryStore{
		items:    make(map[string]memoryItem),
		stopChan: make(chan struct{}),
	}

	// Concurrency pattern: Start background worker to prevent stale memory bloat
	go s.startCleaner(cleanupInterval)

	return s
}

// Get retrieves cached item. Performs fast passive expiration check.
func (s *MemoryStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	it, exists := s.items[key]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrCacheMiss
	}

	// Check if item has expired
	if time.Now().After(it.expiresAt) {
		// Passive eviction: Delete expired item immediately on read
		s.mu.Lock()
		if current, ok := s.items[key]; ok && time.Now().After(current.expiresAt) {
			delete(s.items, key)
		}
		s.mu.Unlock()
		return nil, ErrCacheMiss
	}

	// Return a copy to prevent external mutation from altering cached state
	cp := make([]byte, len(it.value))
	copy(cp, it.value)
	return cp, nil
}

// Set stores raw bytes with specified TTL.
func (s *MemoryStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return ErrInvalidTTL
	}

	// Deep copy value
	cp := make([]byte, len(value))
	copy(cp, value)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = memoryItem{
		value:     cp,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// Delete removes a single key from cache.
func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

// DeletePrefix removes all keys matching prefix (e.g. "users:*").
func (s *MemoryStore) DeletePrefix(ctx context.Context, prefix string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k := range s.items {
		if strings.HasPrefix(k, prefix) {
			delete(s.items, k)
		}
	}
	return nil
}

// Clear flushes entire memory store.
func (s *MemoryStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]memoryItem)
	return nil
}

// Size returns total number of entries currently stored (including any not yet evicted).
func (s *MemoryStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// startCleaner runs periodically to sweep and evict expired items from RAM.
func (s *MemoryStore) startCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.evictExpired()
		case <-s.stopChan:
			return // Gracefully terminates goroutine when store is closed
		}
	}
}

// evictExpired removes expired records.
func (s *MemoryStore) evictExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, it := range s.items {
		if now.After(it.expiresAt) {
			delete(s.items, k)
		}
	}
}

// Close cleanly stops the background cleanup goroutine.
func (s *MemoryStore) Close() error {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
	return nil
}
