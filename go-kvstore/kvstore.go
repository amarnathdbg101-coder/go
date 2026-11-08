package gokvstore

import (
	"hash/fnv"
	"sync"
	"time"
)

type item struct {
	key       string
	val       any
	expiresAt int64 // Unix nanoseconds, 0 = no expiration
}

func (it *item) isExpired(now int64) bool {
	return it.expiresAt > 0 && now >= it.expiresAt
}

type shard struct {
	mu    sync.RWMutex
	items map[string]*item
}

func newShard() *shard {
	return &shard{
		items: make(map[string]*item),
	}
}

// Config specifies store options.
type Config struct {
	ShardCount      int
	CleanupInterval time.Duration
}

// DefaultConfig provides recommended production defaults.
func DefaultConfig() Config {
	return Config{
		ShardCount:      32,
		CleanupInterval: 1 * time.Second,
	}
}

// Store is a high-concurrency sharded in-memory key-value store with TTL.
type Store struct {
	shards    []*shard
	shardMask uint32
	stopChan  chan struct{}
	closed    bool
	closeMu   sync.RWMutex
}

// New initializes the store and starts background active TTL eviction.
func New(cfg Config) *Store {
	if cfg.ShardCount <= 0 {
		cfg.ShardCount = 32
	}
	// Round up to power of 2 for fast bitwise masking
	shardCount := 1
	for shardCount < cfg.ShardCount {
		shardCount <<= 1
	}

	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 1 * time.Second
	}

	s := &Store{
		shards:    make([]*shard, shardCount),
		shardMask: uint32(shardCount - 1),
		stopChan:  make(chan struct{}),
	}

	for i := 0; i < shardCount; i++ {
		s.shards[i] = newShard()
	}

	go s.janitor(cfg.CleanupInterval)

	return s
}

func (s *Store) getShard(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() & s.shardMask
	return s.shards[idx]
}

// Set stores a key-value pair with an optional TTL (0 for infinite).
func (s *Store) Set(key string, val any, ttl time.Duration) {
	var expiresAt int64
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).UnixNano()
	}

	sh := s.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.items[key] = &item{
		key:       key,
		val:       val,
		expiresAt: expiresAt,
	}
}

// Get retrieves a value, returning false if missing or expired.
func (s *Store) Get(key string) (any, bool) {
	sh := s.getShard(key)
	now := time.Now().UnixNano()

	sh.mu.RLock()
	it, ok := sh.items[key]
	if !ok {
		sh.mu.RUnlock()
		return nil, false
	}

	if it.isExpired(now) {
		sh.mu.RUnlock()
		// Lazy eviction
		sh.mu.Lock()
		delete(sh.items, key)
		sh.mu.Unlock()
		return nil, false
	}

	val := it.val
	sh.mu.RUnlock()
	return val, true
}

// GetOrSet returns existing value or sets it if absent or expired.
func (s *Store) GetOrSet(key string, val any, ttl time.Duration) (actual any, loaded bool) {
	sh := s.getShard(key)
	now := time.Now().UnixNano()

	sh.mu.Lock()
	defer sh.mu.Unlock()

	it, ok := sh.items[key]
	if ok && !it.isExpired(now) {
		return it.val, true
	}

	var expiresAt int64
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).UnixNano()
	}

	sh.items[key] = &item{
		key:       key,
		val:       val,
		expiresAt: expiresAt,
	}

	return val, false
}

// CompareAndSwap atomically replaces the value if current value matches oldVal.
func (s *Store) CompareAndSwap(key string, oldVal, newVal any) bool {
	sh := s.getShard(key)
	now := time.Now().UnixNano()

	sh.mu.Lock()
	defer sh.mu.Unlock()

	it, ok := sh.items[key]
	if !ok || it.isExpired(now) {
		return false
	}

	if it.val != oldVal {
		return false
	}

	it.val = newVal
	return true
}

// Increment adds delta to an integer key (creates with delta if non-existent).
func (s *Store) Increment(key string, delta int64) (int64, error) {
	sh := s.getShard(key)
	now := time.Now().UnixNano()

	sh.mu.Lock()
	defer sh.mu.Unlock()

	it, ok := sh.items[key]
	if !ok || it.isExpired(now) {
		sh.items[key] = &item{
			key:       key,
			val:       delta,
			expiresAt: 0,
		}
		return delta, nil
	}

	switch v := it.val.(type) {
	case int64:
		newVal := v + delta
		it.val = newVal
		return newVal, nil
	case int:
		newVal := int64(v) + delta
		it.val = newVal
		return newVal, nil
	default:
		return 0, ErrNotAnInteger
	}
}

// Delete removes a key from the store.
func (s *Store) Delete(key string) bool {
	sh := s.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	_, ok := sh.items[key]
	if ok {
		delete(sh.items, key)
		return true
	}
	return false
}

// Len returns the current count of non-expired keys across all shards.
func (s *Store) Len() int {
	now := time.Now().UnixNano()
	total := 0

	for _, sh := range s.shards {
		sh.mu.RLock()
		for _, it := range sh.items {
			if !it.isExpired(now) {
				total++
			}
		}
		sh.mu.RUnlock()
	}

	return total
}

// Flush clears all keys from all shards.
func (s *Store) Flush() {
	for _, sh := range s.shards {
		sh.mu.Lock()
		sh.items = make(map[string]*item)
		sh.mu.Unlock()
	}
}

func (s *Store) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixNano()
			for _, sh := range s.shards {
				sh.mu.Lock()
				for k, it := range sh.items {
					if it.isExpired(now) {
						delete(sh.items, k)
					}
				}
				sh.mu.Unlock()
			}
		case <-s.stopChan:
			return
		}
	}
}

// Close stops background janitor goroutine.
func (s *Store) Close() {
	s.closeMu.Lock()
	if s.closed {
		s.closeMu.Unlock()
		return
	}
	s.closed = true
	s.closeMu.Unlock()

	close(s.stopChan)
}
