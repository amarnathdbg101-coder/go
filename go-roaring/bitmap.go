package roaring

import (
	"sort"
	"sync"
)

// Bitmap represents a compressed 32-bit Roaring Bitmap.
type Bitmap struct {
	mu         sync.RWMutex
	containers map[uint16]container
}

// New creates a new empty Roaring Bitmap.
func New() *Bitmap {
	return &Bitmap{
		containers: make(map[uint16]container),
	}
}

// Add inserts a 32-bit integer into the bitmap.
func (b *Bitmap) Add(x uint32) {
	high := uint16(x >> 16)
	low := uint16(x & 0xFFFF)

	b.mu.Lock()
	defer b.mu.Unlock()

	c, exists := b.containers[high]
	if !exists {
		c = newArrayContainer()
	}
	b.containers[high] = c.add(low)
}

// AddMany inserts a slice of 32-bit integers into the bitmap.
func (b *Bitmap) AddMany(vals ...uint32) {
	for _, v := range vals {
		b.Add(v)
	}
}

// Remove deletes a 32-bit integer from the bitmap.
func (b *Bitmap) Remove(x uint32) {
	high := uint16(x >> 16)
	low := uint16(x & 0xFFFF)

	b.mu.Lock()
	defer b.mu.Unlock()

	c, exists := b.containers[high]
	if !exists {
		return
	}
	newC := c.remove(low)
	if newC.cardinality() == 0 {
		delete(b.containers, high)
	} else {
		b.containers[high] = newC
	}
}

// Contains checks if the integer is present in the bitmap.
func (b *Bitmap) Contains(x uint32) bool {
	high := uint16(x >> 16)
	low := uint16(x & 0xFFFF)

	b.mu.RLock()
	defer b.mu.RUnlock()

	c, exists := b.containers[high]
	if !exists {
		return false
	}
	return c.contains(low)
}

// Cardinality returns the exact count of unique integers stored in the bitmap.
func (b *Bitmap) Cardinality() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var count uint64
	for _, c := range b.containers {
		count += uint64(c.cardinality())
	}
	return count
}

// ToArray extracts all values stored in the bitmap as an ascending sorted slice of uint32.
func (b *Bitmap) ToArray() []uint32 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Sort high keys
	keys := make([]uint16, 0, len(b.containers))
	for k := range b.containers {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	result := make([]uint32, 0, b.Cardinality())
	for _, k := range keys {
		high := uint32(k) << 16
		for _, low := range b.containers[k].toSlice() {
			result = append(result, high|uint32(low))
		}
	}
	return result
}

// And computes the intersection of two bitmaps (A ∩ B).
func (b *Bitmap) And(other *Bitmap) *Bitmap {
	if other == nil {
		return New()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	res := New()
	for high, c1 := range b.containers {
		if c2, exists := other.containers[high]; exists {
			andC := c1.and(c2)
			if andC.cardinality() > 0 {
				res.containers[high] = andC
			}
		}
	}
	return res
}

// cloneContainer returns an identical copy of the container.
func cloneContainer(c container) container {
	if bc, ok := c.(*bitmapContainer); ok {
		res := newBitmapContainer()
		copy(res.words[:], bc.words[:])
		res.card = bc.card
		return res
	}
	ac := c.(*arrayContainer)
	res := newArrayContainer()
	res.content = make([]uint16, len(ac.content))
	copy(res.content, ac.content)
	return res
}

// Or computes the union of two bitmaps (A ∪ B).
func (b *Bitmap) Or(other *Bitmap) *Bitmap {
	if other == nil {
		return b.Clone()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	res := New()
	for high, c1 := range b.containers {
		if c2, exists := other.containers[high]; exists {
			res.containers[high] = c1.or(c2)
		} else {
			res.containers[high] = cloneContainer(c1)
		}
	}

	for high, c2 := range other.containers {
		if _, exists := b.containers[high]; !exists {
			res.containers[high] = cloneContainer(c2)
		}
	}
	return res
}

// AndNot computes the difference of two bitmaps (A \ B).
func (b *Bitmap) AndNot(other *Bitmap) *Bitmap {
	if other == nil {
		return b.Clone()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	res := New()
	for high, c1 := range b.containers {
		if c2, exists := other.containers[high]; exists {
			diffC := c1.andNot(c2)
			if diffC.cardinality() > 0 {
				res.containers[high] = diffC
			}
		} else {
			res.containers[high] = cloneContainer(c1)
		}
	}
	return res
}

// Clone creates a deep copy of the bitmap.
func (b *Bitmap) Clone() *Bitmap {
	b.mu.RLock()
	defer b.mu.RUnlock()

	clone := New()
	for k, c := range b.containers {
		clone.containers[k] = cloneContainer(c)
	}
	return clone
}
