package sstable

import (
	"hash/fnv"
)

type bloomFilter struct {
	bits []byte
	k    uint8
}

func newBloomFilter(keys [][]byte, bitsPerKey int) *bloomFilter {
	if bitsPerKey < 1 {
		bitsPerKey = 10
	}
	n := len(keys)
	if n == 0 {
		return &bloomFilter{bits: make([]byte, 8), k: 3}
	}

	numBits := n * bitsPerKey
	if numBits < 64 {
		numBits = 64
	}
	// Round up to bytes
	numBytes := (numBits + 7) / 8
	numBits = numBytes * 8

	// Optimal k = (m/n) * ln(2) ~= bitsPerKey * 0.69
	k := uint8(float64(bitsPerKey) * 0.69)
	if k < 1 {
		k = 1
	} else if k > 30 {
		k = 30
	}

	filter := &bloomFilter{
		bits: make([]byte, numBytes),
		k:    k,
	}

	for _, key := range keys {
		filter.add(key, numBits)
	}

	return filter
}

func (b *bloomFilter) add(key []byte, numBits int) {
	h1, h2 := hashPair(key)
	for i := uint8(0); i < b.k; i++ {
		combined := (h1 + uint64(i)*h2) % uint64(numBits)
		byteIdx := combined / 8
		bitIdx := combined % 8
		b.bits[byteIdx] |= (1 << bitIdx)
	}
}

func (b *bloomFilter) contains(key []byte) bool {
	if len(b.bits) == 0 {
		return false
	}
	numBits := len(b.bits) * 8
	h1, h2 := hashPair(key)
	for i := uint8(0); i < b.k; i++ {
		combined := (h1 + uint64(i)*h2) % uint64(numBits)
		byteIdx := combined / 8
		bitIdx := combined % 8
		if (b.bits[byteIdx] & (1 << bitIdx)) == 0 {
			return false
		}
	}
	return true
}

func hashPair(data []byte) (uint64, uint64) {
	h := fnv.New64a()
	_, _ = h.Write(data)
	h1 := h.Sum64()

	// Secondary hash using bit mix
	h2 := h1>>32 | h1<<32
	if h2%2 == 0 {
		h2++
	}
	return h1, h2
}
