package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Rule defines rate limit parameters for an endpoint or router.
type Rule struct {
	// Capacity is the maximum number of requests allowed within the window duration.
	Capacity int

	// Window is the time period over which requests are measured (e.g., 1 minute).
	Window time.Duration
}

// Result contains rate limit evaluation metrics returned to HTTP middleware.
type Result struct {
	// Allowed is true if the request is within capacity, false if rate limit exceeded.
	Allowed bool

	// Remaining is the number of tokens left in the current window.
	Remaining int

	// ResetAt is the UTC time when full capacity will be restored.
	ResetAt time.Time

	// RetryAfter is the recommended waiting duration before making another request.
	RetryAfter time.Duration
}

// Store defines the storage engine contract for go-ratelimit.
// Allows plugging in In-Memory, Redis, Memcached, or custom storage backends.
type Store interface {
	// Take evaluates rate limit state for key and consumes 1 unit of capacity if allowed.
	Take(ctx context.Context, key string, rule Rule) (Result, error)

	// Close gracefully stops background routines and releases resources.
	Close() error
}

// -------------------------------------------------------------------------
// In-Memory Token Bucket Implementation (Zero External Dependencies)
// -------------------------------------------------------------------------

// tokenBucket represents an active rate limiter state for a single key.
type tokenBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// MemoryStore is an ultra-fast, thread-safe, in-memory Token Bucket Store.
// Features 0ms evaluation speed and automatic background cleanup of stale keys.
type MemoryStore struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewMemoryStore creates a new MemoryStore and starts background key cleanup.
func NewMemoryStore() *MemoryStore {
	m := &MemoryStore{
		buckets:  make(map[string]*tokenBucket),
		stopChan: make(chan struct{}),
	}
	go m.startCleanupLoop()
	return m
}

// Take implements Token Bucket evaluation algorithm.
func (m *MemoryStore) Take(ctx context.Context, key string, rule Rule) (Result, error) {
	if rule.Capacity <= 0 || rule.Window <= 0 {
		return Result{Allowed: true, Remaining: rule.Capacity, ResetAt: time.Now()}, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	b, exists := m.buckets[key]

	if !exists {
		b = &tokenBucket{
			tokens:     float64(rule.Capacity - 1), // Consume 1 token for current request
			lastUpdate: now,
		}
		m.buckets[key] = b

		return Result{
			Allowed:    true,
			Remaining:  rule.Capacity - 1,
			ResetAt:    now.Add(rule.Window),
			RetryAfter: 0,
		}, nil
	}

	// Calculate refilled tokens based on elapsed time
	elapsed := now.Sub(b.lastUpdate)
	b.lastUpdate = now

	// Refill rate: tokensPerSecond = rule.Capacity / rule.Window.Seconds()
	refillRate := float64(rule.Capacity) / rule.Window.Seconds()
	b.tokens += elapsed.Seconds() * refillRate

	if b.tokens > float64(rule.Capacity) {
		b.tokens = float64(rule.Capacity)
	}

	// Check if 1 token can be consumed
	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		remaining := int(b.tokens)

		return Result{
			Allowed:    true,
			Remaining:  remaining,
			ResetAt:    now.Add(rule.Window),
			RetryAfter: 0,
		}, nil
	}

	// Capacity exceeded
	neededTokens := 1.0 - b.tokens
	retrySeconds := neededTokens / refillRate
	retryAfter := time.Duration(retrySeconds * float64(time.Second))

	return Result{
		Allowed:    false,
		Remaining:  0,
		ResetAt:    now.Add(retryAfter),
		RetryAfter: retryAfter,
	}, nil
}

// startCleanupLoop runs periodically to delete inactive buckets and prevent memory leaks.
func (m *MemoryStore) startCleanupLoop() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupStaleBuckets()
		case <-m.stopChan:
			return
		}
	}
}

// cleanupStaleBuckets removes buckets that haven't been accessed for more than 10 minutes.
func (m *MemoryStore) cleanupStaleBuckets() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, bucket := range m.buckets {
		if now.Sub(bucket.lastUpdate) > 10*time.Minute {
			delete(m.buckets, key)
		}
	}
}

// Close gracefully stops the background cleanup loop.
func (m *MemoryStore) Close() error {
	m.stopOnce.Do(func() {
		close(m.stopChan)
	})
	return nil
}
