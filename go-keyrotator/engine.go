package keyrotator

import (
	"strings"
	"sync"
	"time"
)

// Rotator manages thread-safe credential pooling, health telemetry, and automatic failovers.
type Rotator struct {
	mu           sync.RWMutex
	keys         []string
	keyHealths  []*KeyHealth
	health       map[string]*KeyHealth
	currentIndex int
	strategy     Strategy
	cfg          *Config
}

// New creates an initialized Rotator with cleaned keys and functional options.
func New(keys []string, opts ...Option) *Rotator {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	cleanKeys := make([]string, 0, len(keys))
	keyHealths := make([]*KeyHealth, 0, len(keys))
	healthMap := make(map[string]*KeyHealth)

	for _, k := range keys {
		clean := strings.TrimSpace(strings.Trim(k, `"'`))
		if clean != "" {
			cleanKeys = append(cleanKeys, clean)
			h := &KeyHealth{
				Key: clean,
			}
			keyHealths = append(keyHealths, h)
			healthMap[clean] = h
		}
	}

	return &Rotator{
		keys:         cleanKeys,
		keyHealths:  keyHealths,
		health:       healthMap,
		currentIndex: 0,
		strategy:     NewRoundRobinStrategy(),
		cfg:          cfg,
	}
}

// KeyCount returns total keys in pool.
func (r *Rotator) KeyCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.keys)
}

// HasKeys returns true if at least one key is loaded.
func (r *Rotator) HasKeys() bool {
	return r.KeyCount() > 0
}

// GetNextKey selects the next available non-rate-limited key.
// High-performance hot path: achieves 0 heap allocations (0 B/op) via pre-allocated cyclic indexing.
// If all keys are in cooldown, it returns the key that will recover earliest along with ErrAllKeysExhausted.
func (r *Rotator) GetNextKey() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	total := len(r.keyHealths)
	if total == 0 {
		return "", ErrNoKeysConfigured
	}

	now := time.Now()

	// 1. Direct cyclic probe across pre-allocated slice without heap allocation
	for i := 0; i < total; i++ {
		idx := (r.currentIndex + i) % total
		h := r.keyHealths[idx]

		// Check if cooldown has elapsed
		if h.IsRateLimited && now.After(h.RateLimitedUntil) {
			h.IsRateLimited = false
			h.ConsecutiveErrors = 0
			if r.cfg.Hooks.OnCooldownRecovered != nil {
				r.cfg.Hooks.OnCooldownRecovered(h.Key)
			}
		}

		if !h.IsRateLimited {
			r.currentIndex = (idx + 1) % total
			h.TotalCalls++
			h.LastUsedAt = now
			if r.cfg.Hooks.OnKeySelected != nil {
				r.cfg.Hooks.OnKeySelected(h.Key)
			}
			return h.Key, nil
		}
	}

	// 2. Fallback: Find key recovering earliest (zero allocations)
	var earliestKey string
	var earliestTime time.Time

	for i := 0; i < total; i++ {
		h := r.keyHealths[i]
		if earliestKey == "" || h.RateLimitedUntil.Before(earliestTime) {
			earliestKey = h.Key
			earliestTime = h.RateLimitedUntil
		}
	}

	return earliestKey, ErrAllKeysExhausted
}

// MarkRateLimited puts a key into temporary quarantine (e.g. upon receiving HTTP 429).
func (r *Rotator) MarkRateLimited(key string, cooldown time.Duration) {
	if cooldown <= 0 {
		cooldown = r.cfg.DefaultCooldown
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if h, exists := r.health[key]; exists {
		h.IsRateLimited = true
		h.RateLimitedUntil = time.Now().Add(cooldown)
		h.ConsecutiveErrors++
		h.TotalErrors++
		if r.cfg.Hooks.OnRateLimited != nil {
			r.cfg.Hooks.OnRateLimited(key, cooldown)
		}
	}
}

// MarkSuccess records successful execution, resetting error streaks.
func (r *Rotator) MarkSuccess(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if h, exists := r.health[key]; exists {
		h.ConsecutiveErrors = 0
		h.IsRateLimited = false
		h.TotalSuccess++
	}
}

// MarkError records failure. If error threshold is breached, the key enters automatic cooldown.
func (r *Rotator) MarkError(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if h, exists := r.health[key]; exists {
		h.ConsecutiveErrors++
		h.TotalErrors++
		if r.cfg.Hooks.OnErrorTriggered != nil {
			r.cfg.Hooks.OnErrorTriggered(key, h.ConsecutiveErrors)
		}
		if h.ConsecutiveErrors >= r.cfg.ErrorThreshold {
			h.IsRateLimited = true
			h.RateLimitedUntil = time.Now().Add(r.cfg.ErrorCooldown)
			if r.cfg.Hooks.OnRateLimited != nil {
				r.cfg.Hooks.OnRateLimited(key, r.cfg.ErrorCooldown)
			}
		}
	}
}

// GetHealth returns a copy of telemetry metrics for a key.
func (r *Rotator) GetHealth(key string) (*KeyHealth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, exists := r.health[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	copy := *h
	return &copy, nil
}

// ListHealth returns snapshots of health metrics for all keys.
func (r *Rotator) ListHealth() []*KeyHealth {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*KeyHealth, 0, len(r.keyHealths))
	for _, h := range r.keyHealths {
		copy := *h
		out = append(out, &copy)
	}
	return out
}
