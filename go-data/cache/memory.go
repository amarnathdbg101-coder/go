package cache

import (
	"context"
	"strings"
	"sync"
	"time"
)

type memoryItem struct {
	value     []byte
	expiresAt time.Time
}

type MemoryStore struct {
	mu       sync.RWMutex
	items    map[string]memoryItem
	stopChan chan struct{}
	stopOnce sync.Once
}

type MemoryOptions struct {
	CleanupInterval time.Duration
}

func NewMemoryStore(opts ...MemoryOptions) *MemoryStore {
	cleanupInterval := 1 * time.Minute
	if len(opts) > 0 && opts[0].CleanupInterval > 0 {
		cleanupInterval = opts[0].CleanupInterval
	}

	s := &MemoryStore{
		items:    make(map[string]memoryItem),
		stopChan: make(chan struct{}),
	}

	go s.startCleaner(cleanupInterval)

	return s
}

func (s *MemoryStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	it, exists := s.items[key]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrCacheMiss
	}

	if time.Now().After(it.expiresAt) {
		s.mu.Lock()
		if current, ok := s.items[key]; ok && time.Now().After(current.expiresAt) {
			delete(s.items, key)
		}
		s.mu.Unlock()
		return nil, ErrCacheMiss
	}

	cp := make([]byte, len(it.value))
	copy(cp, it.value)
	return cp, nil
}

func (s *MemoryStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return ErrInvalidTTL
	}

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

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

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

func (s *MemoryStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]memoryItem)
	return nil
}

func (s *MemoryStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *MemoryStore) startCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
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

func (s *MemoryStore) Close() error {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
	return nil
}
