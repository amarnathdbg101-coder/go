package heavyhitters

import (
	"sort"
	"sync"
)

// Version defines the release version of go-heavyhitters.
const Version = "1.0.0"

// Item represents a detected heavy hitter item with its frequency estimate.
type Item struct {
	Key   string
	Count uint64
	Error uint64
}

type entry struct {
	key   string
	count uint64
	err   uint64
}

// Tracker implements the Space-Saving streaming Heavy Hitters algorithm.
type Tracker struct {
	mu       sync.RWMutex
	capacity int
	total    uint64
	entries  map[string]*entry
}

// New creates an initialized Space-Saving Heavy Hitters tracker.
func New(opts ...Option) (*Tracker, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &Tracker{
		capacity: cfg.Capacity,
		entries:  make(map[string]*entry, cfg.Capacity),
	}, nil
}

// Add records a single occurrence of the item in the stream.
func (t *Tracker) Add(key string) {
	t.AddWithCount(key, 1)
}

// AddWithCount increments the frequency of the item by count.
func (t *Tracker) AddWithCount(key string, count uint64) {
	if count == 0 {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.total += count

	// 1. If key is already monitored, increment count
	if e, exists := t.entries[key]; exists {
		e.count += count
		return
	}

	// 2. If space is available, add new entry
	if len(t.entries) < t.capacity {
		t.entries[key] = &entry{
			key:   key,
			count: count,
			err:   0,
		}
		return
	}

	// 3. Table is full: find element with minimum count
	var minKey string
	var minCount uint64 = ^uint64(0)
	var minErr uint64

	for k, e := range t.entries {
		if e.count < minCount {
			minCount = e.count
			minErr = e.err
			minKey = k
		}
	}

	// Evict minKey and replace with new key
	delete(t.entries, minKey)
	t.entries[key] = &entry{
		key:   key,
		count: minCount + count,
		err:   minCount,
	}
	_ = minErr
}

// TopK returns the top K most frequent items sorted in descending order of frequency.
func (t *Tracker) TopK(k int) []Item {
	if k <= 0 {
		return nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	list := make([]Item, 0, len(t.entries))
	for _, e := range t.entries {
		list = append(list, Item{
			Key:   e.key,
			Count: e.count,
			Error: e.err,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Count != list[j].Count {
			return list[i].Count > list[j].Count
		}
		return list[i].Key < list[j].Key
	})

	if k < len(list) {
		return list[:k]
	}
	return list
}

// Total returns the total number of items observed in the stream.
func (t *Tracker) Total() uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.total
}

// Reset clears all counters.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = make(map[string]*entry, t.capacity)
	t.total = 0
}
