package ratelimit

import (
	"context"
	"sync"
	"time"
)

type Rule struct {
	Capacity int
	Window   time.Duration
}

type Result struct {
	Allowed    bool
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}

type Store interface {
	Take(ctx context.Context, key string, rule Rule) (Result, error)
	Close() error
}

type tokenBucket struct {
	tokens     float64
	lastUpdate time.Time
}

type MemoryStore struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	stopChan chan struct{}
	stopOnce sync.Once
}

func NewMemoryStore() *MemoryStore {
	m := &MemoryStore{
		buckets:  make(map[string]*tokenBucket),
		stopChan: make(chan struct{}),
	}
	go m.startCleanupLoop()
	return m
}

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
			tokens:     float64(rule.Capacity - 1),
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

	elapsed := now.Sub(b.lastUpdate)
	b.lastUpdate = now

	refillRate := float64(rule.Capacity) / rule.Window.Seconds()
	b.tokens += elapsed.Seconds() * refillRate

	if b.tokens > float64(rule.Capacity) {
		b.tokens = float64(rule.Capacity)
	}

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

func (m *MemoryStore) Close() error {
	m.stopOnce.Do(func() {
		close(m.stopChan)
	})
	return nil
}
