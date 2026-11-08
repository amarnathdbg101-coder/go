package countingbloom

import (
	"hash/fnv"
	"math"
	"sync"
)

// Filter represents a Counting Bloom Filter supporting additions and deletions.
type Filter struct {
	mu         sync.RWMutex
	concurrent bool
	counters   []uint8
	numHashes  uint
	size       uint
	count      uint64
}

// New creates a new Counting Bloom Filter with optimal sizing.
func New(expectedElements uint, falsePositiveRate float64) (*Filter, error) {
	if expectedElements == 0 {
		return nil, ErrInvalidCapacity
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, ErrInvalidFPRate
	}

	// Calculate optimal filter size M and hash count K
	n := float64(expectedElements)
	p := falsePositiveRate

	m := math.Ceil(-1 * (n * math.Log(p)) / math.Pow(math.Log(2), 2))
	k := math.Round((m / n) * math.Log(2))

	if k < 1 {
		k = 1
	}

	size := uint(m)
	return &Filter{
		counters:  make([]uint8, size),
		numHashes: uint(k),
		size:      size,
	}, nil
}

// NewConcurrent creates a thread-safe Counting Bloom Filter.
func NewConcurrent(expectedElements uint, falsePositiveRate float64) (*Filter, error) {
	f, err := New(expectedElements, falsePositiveRate)
	if err != nil {
		return nil, err
	}
	f.concurrent = true
	return f, nil
}

// hashValues computes two 64-bit base hashes for double hashing.
func (f *Filter) hashValues(data []byte) (uint64, uint64) {
	h := fnv.New64a()
	_, _ = h.Write(data)
	h1 := h.Sum64()

	// SplitMix64 variation for independent second hash
	z := h1 + 0x9e3779b97f4a7c15
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	h2 := z ^ (z >> 31)

	return h1, h2
}

func (f *Filter) getIndexes(data []byte) []uint {
	h1, h2 := f.hashValues(data)
	indices := make([]uint, f.numHashes)
	for i := uint(0); i < f.numHashes; i++ {
		indices[i] = uint((h1 + uint64(i)*h2) % uint64(f.size))
	}
	return indices
}

// Add inserts data into the filter, incrementing counters up to saturation 255.
func (f *Filter) Add(data []byte) {
	if f.concurrent {
		f.mu.Lock()
		defer f.mu.Unlock()
	}

	indices := f.getIndexes(data)
	for _, idx := range indices {
		if f.counters[idx] < 255 {
			f.counters[idx]++
		}
	}
	f.count++
}

// AddString inserts a string into the filter.
func (f *Filter) AddString(s string) {
	f.Add([]byte(s))
}

// Contains checks if the data is likely present in the filter.
func (f *Filter) Contains(data []byte) bool {
	if f.concurrent {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}

	indices := f.getIndexes(data)
	for _, idx := range indices {
		if f.counters[idx] == 0 {
			return false
		}
	}
	return true
}

// ContainsString checks if a string is likely present in the filter.
func (f *Filter) ContainsString(s string) bool {
	return f.Contains([]byte(s))
}

// Remove decrements the counters for data.
// Returns true if the element was present and removed, false if definitely not present.
func (f *Filter) Remove(data []byte) bool {
	if f.concurrent {
		f.mu.Lock()
		defer f.mu.Unlock()
	}

	indices := f.getIndexes(data)

	// Verify all counters are > 0 first
	for _, idx := range indices {
		if f.counters[idx] == 0 {
			return false
		}
	}

	// Decrement non-saturated counters
	for _, idx := range indices {
		if f.counters[idx] < 255 { // saturated counters remain to prevent false negatives
			f.counters[idx]--
		}
	}

	if f.count > 0 {
		f.count--
	}
	return true
}

// RemoveString removes a string from the filter.
func (f *Filter) RemoveString(s string) bool {
	return f.Remove([]byte(s))
}

// Count returns the number of active items inserted.
func (f *Filter) Count() uint64 {
	if f.concurrent {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}
	return f.count
}

// Clear resets all counters back to zero.
func (f *Filter) Clear() {
	if f.concurrent {
		f.mu.Lock()
		defer f.mu.Unlock()
	}
	for i := range f.counters {
		f.counters[i] = 0
	}
	f.count = 0
}
