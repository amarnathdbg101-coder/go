package gobloom

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"sync"
)

// Filter is a space-efficient probabilistic set membership structure.
type Filter struct {
	mu       sync.RWMutex
	m        uint64   // Number of bits in filter
	k        uint32   // Number of hash functions
	bits     []uint64 // Bit array packed into 64-bit words
	elements uint64   // Approximate count of items added
}

// New constructs an optimal Bloom Filter sized for expectedItems with the desired falsePositiveRate (e.g. 0.01 for 1%).
func New(expectedItems uint64, falsePositiveRate float64) (*Filter, error) {
	if expectedItems == 0 || falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, ErrInvalidParams
	}

	// m = - (n * ln(p)) / (ln(2)^2)
	n := float64(expectedItems)
	ln2 := math.Ln2
	mFloat := -(n * math.Log(falsePositiveRate)) / (ln2 * ln2)
	m := uint64(math.Ceil(mFloat))

	// k = (m / n) * ln(2)
	kFloat := (mFloat / n) * ln2
	k := uint32(math.Ceil(kFloat))
	if k < 1 {
		k = 1
	}

	wordCount := (m + 63) / 64

	return &Filter{
		m:    m,
		k:    k,
		bits: make([]uint64, wordCount),
	}, nil
}

// Add inserts an item into the filter.
func (f *Filter) Add(item []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()

	h1, h2 := f.hashPair(item)
	for i := uint32(0); i < f.k; i++ {
		idx := (h1 + uint64(i)*h2) % f.m
		word := idx / 64
		bit := idx % 64
		f.bits[word] |= (1 << bit)
	}
	f.elements++
}

// AddString is a convenience helper for string keys.
func (f *Filter) AddString(s string) {
	f.Add([]byte(s))
}

// Contains checks if an item may have been added.
// If returns false, the item is GUARANTEED not in the set.
// If returns true, the item is PROBABLY in the set.
func (f *Filter) Contains(item []byte) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	h1, h2 := f.hashPair(item)
	for i := uint32(0); i < f.k; i++ {
		idx := (h1 + uint64(i)*h2) % f.m
		word := idx / 64
		bit := idx % 64
		if (f.bits[word] & (1 << bit)) == 0 {
			return false
		}
	}
	return true
}

// ContainsString checks if a string item is in the filter.
func (f *Filter) ContainsString(s string) bool {
	return f.Contains([]byte(s))
}

func mix64(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

// hashPair generates two independent 64-bit hashes using Kirsch-Mitzenmacher optimization.
func (f *Filter) hashPair(data []byte) (uint64, uint64) {
	hasher1 := fnv.New64a()
	hasher1.Write(data)
	h1 := hasher1.Sum64()

	hasher2 := fnv.New64()
	hasher2.Write(data)
	h2 := mix64(hasher2.Sum64())

	if h2 == 0 {
		h2 = 1 // Ensure h2 is never zero to avoid degenerating the probe sequence
	}

	return h1, h2
}

// Reset clears all bits in the filter.
func (f *Filter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := range f.bits {
		f.bits[i] = 0
	}
	f.elements = 0
}

// SizeInBytes returns memory footprint of the bitset in bytes.
func (f *Filter) SizeInBytes() int {
	return len(f.bits) * 8
}

// Bytes serializes the Bloom filter for persistence (format: [8B m | 4B k | 8B elements | bitset]).
func (f *Filter) Bytes() []byte {
	f.mu.RLock()
	defer f.mu.RUnlock()

	buf := make([]byte, 20+len(f.bits)*8)
	binary.BigEndian.PutUint64(buf[0:8], f.m)
	binary.BigEndian.PutUint32(buf[8:12], f.k)
	binary.BigEndian.PutUint64(buf[12:20], f.elements)

	for i, word := range f.bits {
		binary.BigEndian.PutUint64(buf[20+i*8:28+i*8], word)
	}

	return buf
}

// FromBytes deserializes a Bloom filter from byte representation.
func FromBytes(data []byte) (*Filter, error) {
	if len(data) < 20 {
		return nil, ErrCorruptData
	}

	m := binary.BigEndian.Uint64(data[0:8])
	k := binary.BigEndian.Uint32(data[8:12])
	elements := binary.BigEndian.Uint64(data[12:20])

	wordCount := (m + 63) / 64
	expectedLen := 20 + int(wordCount)*8
	if len(data) != expectedLen {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrCorruptData, expectedLen, len(data))
	}

	bits := make([]uint64, wordCount)
	for i := 0; i < int(wordCount); i++ {
		bits[i] = binary.BigEndian.Uint64(data[20+i*8 : 28+i*8])
	}

	return &Filter{
		m:        m,
		k:        k,
		bits:     bits,
		elements: elements,
	}, nil
}
