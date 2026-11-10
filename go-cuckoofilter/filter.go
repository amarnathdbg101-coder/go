package cuckoofilter

import (
	"hash/fnv"
	"math/rand"
	"sync"
)

// Version defines the release version of go-cuckoofilter.
const Version = "1.0.0"

const (
	bucketSize = 4 // 4 slots per bucket
)

type bucket [bucketSize]uint16

// Filter represents a concurrent Cuckoo Filter.
type Filter struct {
	mu         sync.RWMutex
	cfg        Config
	buckets    []bucket
	numBuckets uint32
	count      uint
	rnd        *rand.Rand
}

// New creates an initialized Cuckoo Filter.
func New(opts ...Option) (*Filter, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Capacity == 0 {
		return nil, ErrInvalidCapacity
	}
	if cfg.MaxKicks == 0 {
		cfg.MaxKicks = 500
	}

	// Calculate number of buckets as power of two >= Capacity / bucketSize
	target := uint32(cfg.Capacity / bucketSize)
	if target < 16 {
		target = 16
	}

	numBuckets := uint32(1)
	for numBuckets < target {
		numBuckets <<= 1
	}

	return &Filter{
		cfg:        cfg,
		buckets:    make([]bucket, numBuckets),
		numBuckets: numBuckets,
		rnd:        rand.New(rand.NewSource(42)),
	}, nil
}

// Insert adds an item to the filter.
func (f *Filter) Insert(data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	fp, i1, i2 := f.indicesAndFingerprint(data)

	// Try inserting into i1
	if f.insertIntoBucket(i1, fp) {
		f.count++
		return nil
	}

	// Try inserting into i2
	if f.insertIntoBucket(i2, fp) {
		f.count++
		return nil
	}

	// Must perform cuckoo kicking
	curIdx := i1
	if f.rnd.Intn(2) == 1 {
		curIdx = i2
	}
	curFP := fp

	for k := uint(0); k < f.cfg.MaxKicks; k++ {
		// Pick random slot in curIdx
		slot := f.rnd.Intn(bucketSize)
		displaced := f.buckets[curIdx][slot]
		f.buckets[curIdx][slot] = curFP

		// Compute alternate bucket for displaced fingerprint
		curIdx = f.alternateIndex(curIdx, displaced)
		curFP = displaced

		// Try placing displaced item in empty slot
		if f.insertIntoBucket(curIdx, curFP) {
			f.count++
			return nil
		}
	}

	return ErrFilterFull
}

// InsertString inserts a string item into the filter.
func (f *Filter) InsertString(str string) error {
	return f.Insert([]byte(str))
}

// Lookup checks if an item is possibly in the filter.
func (f *Filter) Lookup(data []byte) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	fp, i1, i2 := f.indicesAndFingerprint(data)

	// Check i1
	for s := 0; s < bucketSize; s++ {
		if f.buckets[i1][s] == fp {
			return true
		}
	}

	// Check i2
	for s := 0; s < bucketSize; s++ {
		if f.buckets[i2][s] == fp {
			return true
		}
	}

	return false
}

// LookupString checks if a string item is in the filter.
func (f *Filter) LookupString(str string) bool {
	return f.Lookup([]byte(str))
}

// Delete removes an item from the filter. Returns true if removed, false if not present.
func (f *Filter) Delete(data []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	fp, i1, i2 := f.indicesAndFingerprint(data)

	// Check i1
	for s := 0; s < bucketSize; s++ {
		if f.buckets[i1][s] == fp {
			f.buckets[i1][s] = 0
			f.count--
			return true
		}
	}

	// Check i2
	for s := 0; s < bucketSize; s++ {
		if f.buckets[i2][s] == fp {
			f.buckets[i2][s] = 0
			f.count--
			return true
		}
	}

	return false
}

// DeleteString removes a string item from the filter.
func (f *Filter) DeleteString(str string) bool {
	return f.Delete([]byte(str))
}

// Count returns the number of items stored in the filter.
func (f *Filter) Count() uint {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.count
}

// LoadFactor returns the current utilization of the filter (count / total slots).
func (f *Filter) LoadFactor() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	totalSlots := float64(f.numBuckets * bucketSize)
	return float64(f.count) / totalSlots
}

// Reset clears all buckets in the filter.
func (f *Filter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.buckets {
		f.buckets[i] = bucket{}
	}
	f.count = 0
}

func (f *Filter) insertIntoBucket(idx uint32, fp uint16) bool {
	for s := 0; s < bucketSize; s++ {
		if f.buckets[idx][s] == 0 {
			f.buckets[idx][s] = fp
			return true
		}
	}
	return false
}

func (f *Filter) indicesAndFingerprint(data []byte) (fp uint16, i1, i2 uint32) {
	h := fnv.New64a()
	_, _ = h.Write(data)
	hash := h.Sum64()

	// 16-bit non-zero fingerprint
	fp = uint16(hash & 0xFFFF)
	if fp == 0 {
		fp = 1
	}

	i1 = uint32((hash >> 16)) & (f.numBuckets - 1)
	i2 = f.alternateIndex(i1, fp)
	return fp, i1, i2
}

func (f *Filter) alternateIndex(i uint32, fp uint16) uint32 {
	fpHash := uint32(fp) * 0x5bd1e995 // Murmur-style mixing on fingerprint
	return (i ^ fpHash) & (f.numBuckets - 1)
}
