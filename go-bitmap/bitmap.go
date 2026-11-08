package gobitmap

import (
	"encoding/binary"
	"math/bits"
	"sync"
)

// Bitmap is a thread-safe, high-performance bit array utilizing 64-bit word storage.
type Bitmap struct {
	words []uint64
	mu    sync.RWMutex
}

// New creates a new Bitmap initialized to hold at least initialBits capacity.
func New(initialBits int) *Bitmap {
	if initialBits < 0 {
		initialBits = 0
	}
	wordsCount := (initialBits + 63) / 64
	return &Bitmap{
		words: make([]uint64, wordsCount),
	}
}

// Set marks the specified bit index to 1. Dynamically expands if bit exceeds current capacity.
func (b *Bitmap) Set(bit int) error {
	if bit < 0 {
		return ErrOutOfBounds
	}

	wordIdx := bit >> 6
	bitOffset := uint(bit & 0x3f)

	b.mu.Lock()
	defer b.mu.Unlock()

	if wordIdx >= len(b.words) {
		newWords := make([]uint64, wordIdx+1)
		copy(newWords, b.words)
		b.words = newWords
	}

	b.words[wordIdx] |= (1 << bitOffset)
	return nil
}

// Clear sets the specified bit index to 0.
func (b *Bitmap) Clear(bit int) error {
	if bit < 0 {
		return ErrOutOfBounds
	}

	wordIdx := bit >> 6
	bitOffset := uint(bit & 0x3f)

	b.mu.Lock()
	defer b.mu.Unlock()

	if wordIdx < len(b.words) {
		b.words[wordIdx] &= ^(1 << bitOffset)
	}
	return nil
}

// Toggle flips the specified bit index (0 -> 1, 1 -> 0).
func (b *Bitmap) Toggle(bit int) error {
	if bit < 0 {
		return ErrOutOfBounds
	}

	wordIdx := bit >> 6
	bitOffset := uint(bit & 0x3f)

	b.mu.Lock()
	defer b.mu.Unlock()

	if wordIdx >= len(b.words) {
		newWords := make([]uint64, wordIdx+1)
		copy(newWords, b.words)
		b.words = newWords
	}

	b.words[wordIdx] ^= (1 << bitOffset)
	return nil
}

// Test checks if the specified bit index is set to 1.
func (b *Bitmap) Test(bit int) bool {
	if bit < 0 {
		return false
	}

	wordIdx := bit >> 6
	bitOffset := uint(bit & 0x3f)

	b.mu.RLock()
	defer b.mu.RUnlock()

	if wordIdx >= len(b.words) {
		return false
	}

	return (b.words[wordIdx] & (1 << bitOffset)) != 0
}

// Count returns the total number of set bits (1s) using hardware-accelerated POPCNT.
func (b *Bitmap) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	total := 0
	for _, word := range b.words {
		total += bits.OnesCount64(word)
	}
	return total
}

// And computes the bitwise intersection (b AND other).
func (b *Bitmap) And(other *Bitmap) *Bitmap {
	b.mu.RLock()
	defer b.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	minLen := len(b.words)
	if len(other.words) < minLen {
		minLen = len(other.words)
	}

	result := &Bitmap{
		words: make([]uint64, minLen),
	}
	for i := 0; i < minLen; i++ {
		result.words[i] = b.words[i] & other.words[i]
	}

	return result
}

// Or computes the bitwise union (b OR other).
func (b *Bitmap) Or(other *Bitmap) *Bitmap {
	b.mu.RLock()
	defer b.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	maxLen := len(b.words)
	if len(other.words) > maxLen {
		maxLen = len(other.words)
	}

	result := &Bitmap{
		words: make([]uint64, maxLen),
	}
	for i := 0; i < maxLen; i++ {
		var w1, w2 uint64
		if i < len(b.words) {
			w1 = b.words[i]
		}
		if i < len(other.words) {
			w2 = other.words[i]
		}
		result.words[i] = w1 | w2
	}

	return result
}

// AndNot computes the bitwise difference (b AND NOT other).
func (b *Bitmap) AndNot(other *Bitmap) *Bitmap {
	b.mu.RLock()
	defer b.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	result := &Bitmap{
		words: make([]uint64, len(b.words)),
	}
	for i := 0; i < len(b.words); i++ {
		var w2 uint64
		if i < len(other.words) {
			w2 = other.words[i]
		}
		result.words[i] = b.words[i] & (^w2)
	}

	return result
}

// Reset clears all bits to 0.
func (b *Bitmap) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := range b.words {
		b.words[i] = 0
	}
}

// ToBytes serializes the bitmap to raw bytes.
func (b *Bitmap) ToBytes() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()

	buf := make([]byte, len(b.words)*8)
	for i, w := range b.words {
		binary.LittleEndian.PutUint64(buf[i*8:], w)
	}
	return buf
}

// FromBytes deserializes raw bytes into a Bitmap.
func FromBytes(data []byte) *Bitmap {
	wordsCount := (len(data) + 7) / 8
	bm := &Bitmap{
		words: make([]uint64, wordsCount),
	}

	for i := 0; i < len(data)/8; i++ {
		bm.words[i] = binary.LittleEndian.Uint64(data[i*8:])
	}

	// Handle trailing bytes if not multiple of 8
	if rem := len(data) % 8; rem != 0 {
		var lastWord uint64
		offset := (len(data) / 8) * 8
		for j := 0; j < rem; j++ {
			lastWord |= uint64(data[offset+j]) << (j * 8)
		}
		bm.words[wordsCount-1] = lastWord
	}

	return bm
}
