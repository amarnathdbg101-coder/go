package secguard

import (
	"context"
	"sync"
	"time"
)

// MemoryJail provides an in-memory, thread-safe, bounded IP quarantine storage.
type MemoryJail struct {
	mu          sync.RWMutex
	jailMap     map[string]JailEntry
	maxEntries  int
	stopCleanup chan struct{}
	closeOnce   sync.Once
	isClosed    bool
}

// NewMemoryJail creates a bounded memory jail with a background purge worker.
func NewMemoryJail(maxEntries int, cleanupInterval time.Duration) *MemoryJail {
	if maxEntries <= 0 {
		maxEntries = 10000
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 1 * time.Minute
	}

	j := &MemoryJail{
		jailMap:     make(map[string]JailEntry),
		maxEntries:  maxEntries,
		stopCleanup: make(chan struct{}),
	}

	go j.janitor(cleanupInterval)
	return j
}

func (j *MemoryJail) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.PurgeExpired()
		case <-j.stopCleanup:
			return
		}
	}
}

// PurgeExpired clears expired records to keep memory bounded.
func (j *MemoryJail) PurgeExpired() {
	j.mu.Lock()
	defer j.mu.Unlock()

	now := time.Now()
	for ip, entry := range j.jailMap {
		if now.After(entry.ExpiresAt) {
			delete(j.jailMap, ip)
		}
	}
}

// Jail places an IP into quarantine.
func (j *MemoryJail) Jail(_ context.Context, ip string, duration time.Duration, reason string) error {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return nil
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	if j.isClosed {
		return nil
	}

	// Memory exhaustion defense under heavy flood
	if len(j.jailMap) >= j.maxEntries {
		now := time.Now()
		for k, v := range j.jailMap {
			if now.After(v.ExpiresAt) {
				delete(j.jailMap, k)
			}
		}
		if len(j.jailMap) >= j.maxEntries {
			// Emergency reset to protect server heap
			j.jailMap = make(map[string]JailEntry)
		}
	}

	j.jailMap[ip] = JailEntry{
		IP:        ip,
		Reason:    reason,
		ExpiresAt: time.Now().Add(duration),
	}
	return nil
}

// IsJailed checks if the IP is currently in quarantine.
func (j *MemoryJail) IsJailed(_ context.Context, ip string) (bool, string, int64, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	if j.isClosed {
		return false, "", 0, nil
	}

	rec, exists := j.jailMap[ip]
	if !exists {
		return false, "", 0, nil
	}

	now := time.Now()
	if now.After(rec.ExpiresAt) {
		return false, "", 0, nil
	}

	remaining := int64(time.Until(rec.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}

	return true, rec.Reason, remaining, nil
}

// Unjail clears an IP from quarantine.
func (j *MemoryJail) Unjail(_ context.Context, ip string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.jailMap, ip)
	return nil
}

// Close gracefully stops the cleanup janitor.
func (j *MemoryJail) Close() error {
	j.closeOnce.Do(func() {
		j.mu.Lock()
		j.isClosed = true
		j.jailMap = nil
		j.mu.Unlock()
		close(j.stopCleanup)
	})
	return nil
}
