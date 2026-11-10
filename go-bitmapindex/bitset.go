package bitmapindex

import (
	"math/bits"
)

type bitset struct {
	words []uint64
}

func newBitset() *bitset {
	return &bitset{words: make([]uint64, 0, 8)}
}

func (b *bitset) set(i uint32) {
	wordIdx := int(i / 64)
	bitIdx := i % 64

	if wordIdx >= len(b.words) {
		newWords := make([]uint64, wordIdx+1)
		copy(newWords, b.words)
		b.words = newWords
	}
	b.words[wordIdx] |= (uint64(1) << bitIdx)
}

func (b *bitset) test(i uint32) bool {
	wordIdx := int(i / 64)
	if wordIdx >= len(b.words) {
		return false
	}
	return (b.words[wordIdx] & (uint64(1) << (i % 64))) != 0
}

func (b *bitset) count() uint64 {
	var c uint64
	for _, w := range b.words {
		c += uint64(bits.OnesCount64(w))
	}
	return c
}

func (b *bitset) and(other *bitset) *bitset {
	minLen := len(b.words)
	if len(other.words) < minLen {
		minLen = len(other.words)
	}
	res := &bitset{words: make([]uint64, minLen)}
	for i := 0; i < minLen; i++ {
		res.words[i] = b.words[i] & other.words[i]
	}
	return res
}

func (b *bitset) toSlice() []uint32 {
	var res []uint32
	for i, w := range b.words {
		base := uint32(i * 64)
		for w != 0 {
			t := w & -w
			bit := uint32(bits.TrailingZeros64(w))
			res = append(res, base+bit)
			w ^= t
		}
	}
	return res
}
