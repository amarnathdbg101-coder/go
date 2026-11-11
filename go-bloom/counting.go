package gobloom

import (
	"hash/fnv"
	"math"
	"sync"
)

// CountingFilter is a Counting Bloom Filter variant that supports both additions and deletions.
// Instead of single-bit flags, it uses 8-bit saturating counters.
type CountingFilter struct {
	mu       sync.RWMutex
	m        uint64
	k        uint32
	counters []uint8
	elements uint64
}

// NewCounting constructs an optimal Counting Bloom Filter supporting additions and safe removals.
func NewCounting(expectedItems uint64, falsePositiveRate float64) (*CountingFilter, error) {
	if expectedItems == 0 || falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, ErrInvalidParams
	}

	n := float64(expectedItems)
	ln2 := math.Ln2
	mFloat := -(n * math.Log(falsePositiveRate)) / (ln2 * ln2)
	m := uint64(math.Ceil(mFloat))

	kFloat := (mFloat / n) * ln2
	k := uint32(math.Ceil(kFloat))
	if k < 1 {
		k = 1
	}

	return &CountingFilter{
		m:        m,
		k:        k,
		counters: make([]uint8, m),
	}, nil
}

// Add increments counters for the item. Counters saturate at 255 to prevent overflow.
func (cf *CountingFilter) Add(item []byte) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	h1, h2 := cf.hashPair(item)
	for i := uint32(0); i < cf.k; i++ {
		idx := (h1 + uint64(i)*h2) % cf.m
		if cf.counters[idx] < 255 {
			cf.counters[idx]++
		}
	}
	cf.elements++
}

// AddString is a convenience wrapper for string elements.
func (cf *CountingFilter) AddString(s string) {
	cf.Add([]byte(s))
}

// Contains checks if the item is probably in the set.
func (cf *CountingFilter) Contains(item []byte) bool {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	h1, h2 := cf.hashPair(item)
	for i := uint32(0); i < cf.k; i++ {
		idx := (h1 + uint64(i)*h2) % cf.m
		if cf.counters[idx] == 0 {
			return false
		}
	}
	return true
}

// ContainsString checks if string is in the filter.
func (cf *CountingFilter) ContainsString(s string) bool {
	return cf.Contains([]byte(s))
}

// Remove decrements counters for an existing item.
// Returns false if any counter was already zero (element was definitely not present).
func (cf *CountingFilter) Remove(item []byte) bool {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	h1, h2 := cf.hashPair(item)
	// First check if all counters are positive
	for i := uint32(0); i < cf.k; i++ {
		idx := (h1 + uint64(i)*h2) % cf.m
		if cf.counters[idx] == 0 {
			return false
		}
	}

	// Decrement non-saturated counters
	for i := uint32(0); i < cf.k; i++ {
		idx := (h1 + uint64(i)*h2) % cf.m
		if cf.counters[idx] > 0 && cf.counters[idx] < 255 {
			cf.counters[idx]--
		}
	}
	if cf.elements > 0 {
		cf.elements--
	}
	return true
}

// RemoveString removes a string from the counting bloom filter.
func (cf *CountingFilter) RemoveString(s string) bool {
	return cf.Remove([]byte(s))
}

// Elements returns approximate count of inserted items.
func (cf *CountingFilter) Elements() uint64 {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.elements
}

// Reset clears all counters back to zero.
func (cf *CountingFilter) Reset() {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	for i := range cf.counters {
		cf.counters[i] = 0
	}
	cf.elements = 0
}

func (cf *CountingFilter) hashPair(data []byte) (uint64, uint64) {
	hasher1 := fnv.New64a()
	hasher1.Write(data)
	h1 := hasher1.Sum64()

	hasher2 := fnv.New64()
	hasher2.Write(data)
	h2 := mix64(hasher2.Sum64())

	if h2 == 0 {
		h2 = 1
	}
	return h1, h2
}
