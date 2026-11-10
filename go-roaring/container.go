package roaring

import (
	"math/bits"
	"sort"
)

const arrayThreshold = 4096

type container interface {
	add(x uint16) container
	remove(x uint16) container
	contains(x uint16) bool
	cardinality() int
	toSlice() []uint16
	and(other container) container
	or(other container) container
	andNot(other container) container
}

// -------------------------------------------------------------
// Array Container (Sparse: < 4096 elements)
// -------------------------------------------------------------

type arrayContainer struct {
	content []uint16
}

func newArrayContainer() *arrayContainer {
	return &arrayContainer{content: make([]uint16, 0, 8)}
}

func (a *arrayContainer) contains(x uint16) bool {
	idx := sort.Search(len(a.content), func(i int) bool {
		return a.content[i] >= x
	})
	return idx < len(a.content) && a.content[idx] == x
}

func (a *arrayContainer) add(x uint16) container {
	idx := sort.Search(len(a.content), func(i int) bool {
		return a.content[i] >= x
	})
	if idx < len(a.content) && a.content[idx] == x {
		return a // already present
	}

	// Insert sorted
	a.content = append(a.content, 0)
	copy(a.content[idx+1:], a.content[idx:])
	a.content[idx] = x

	// Upgrade to bitmap container if threshold reached
	if len(a.content) >= arrayThreshold {
		bc := newBitmapContainer()
		for _, v := range a.content {
			bc.add(v)
		}
		return bc
	}
	return a
}

func (a *arrayContainer) remove(x uint16) container {
	idx := sort.Search(len(a.content), func(i int) bool {
		return a.content[i] >= x
	})
	if idx < len(a.content) && a.content[idx] == x {
		a.content = append(a.content[:idx], a.content[idx+1:]...)
	}
	return a
}

func (a *arrayContainer) cardinality() int {
	return len(a.content)
}

func (a *arrayContainer) toSlice() []uint16 {
	cp := make([]uint16, len(a.content))
	copy(cp, a.content)
	return cp
}

func (a *arrayContainer) and(other container) container {
	res := newArrayContainer()
	for _, v := range a.content {
		if other.contains(v) {
			res.content = append(res.content, v)
		}
	}
	return res
}

func (a *arrayContainer) or(other container) container {
	// If other is bitmap, let bitmap handle union
	if bc, ok := other.(*bitmapContainer); ok {
		return bc.or(a)
	}

	ac := other.(*arrayContainer)
	res := newArrayContainer()
	i, j := 0, 0
	for i < len(a.content) && j < len(ac.content) {
		if a.content[i] < ac.content[j] {
			res.content = append(res.content, a.content[i])
			i++
		} else if a.content[i] > ac.content[j] {
			res.content = append(res.content, ac.content[j])
			j++
		} else {
			res.content = append(res.content, a.content[i])
			i++
			j++
		}
	}
	for i < len(a.content) {
		res.content = append(res.content, a.content[i])
		i++
	}
	for j < len(ac.content) {
		res.content = append(res.content, ac.content[j])
		j++
	}

	if len(res.content) >= arrayThreshold {
		bc := newBitmapContainer()
		for _, v := range res.content {
			bc.add(v)
		}
		return bc
	}
	return res
}

func (a *arrayContainer) andNot(other container) container {
	res := newArrayContainer()
	for _, v := range a.content {
		if !other.contains(v) {
			res.content = append(res.content, v)
		}
	}
	return res
}

// -------------------------------------------------------------
// Bitmap Container (Dense: >= 4096 elements)
// -------------------------------------------------------------

type bitmapContainer struct {
	words [1024]uint64
	card  int
}

func newBitmapContainer() *bitmapContainer {
	return &bitmapContainer{}
}

func (b *bitmapContainer) contains(x uint16) bool {
	wordIdx := x / 64
	bitIdx := x % 64
	return (b.words[wordIdx] & (uint64(1) << bitIdx)) != 0
}

func (b *bitmapContainer) add(x uint16) container {
	wordIdx := x / 64
	bitIdx := x % 64
	mask := uint64(1) << bitIdx
	if (b.words[wordIdx] & mask) == 0 {
		b.words[wordIdx] |= mask
		b.card++
	}
	return b
}

func (b *bitmapContainer) remove(x uint16) container {
	wordIdx := x / 64
	bitIdx := x % 64
	mask := uint64(1) << bitIdx
	if (b.words[wordIdx] & mask) != 0 {
		b.words[wordIdx] &^= mask
		b.card--

		// Downgrade to array container if drops below threshold
		if b.card < arrayThreshold {
			ac := newArrayContainer()
			ac.content = b.toSlice()
			return ac
		}
	}
	return b
}

func (b *bitmapContainer) cardinality() int {
	if b.card == 0 {
		c := 0
		for _, w := range b.words {
			c += bits.OnesCount64(w)
		}
		b.card = c
	}
	return b.card
}

func (b *bitmapContainer) toSlice() []uint16 {
	slice := make([]uint16, 0, b.cardinality())
	for i, w := range b.words {
		base := uint16(i * 64)
		for w != 0 {
			t := w & -w
			bit := uint16(bits.TrailingZeros64(w))
			slice = append(slice, base+bit)
			w ^= t
		}
	}
	return slice
}

func (b *bitmapContainer) and(other container) container {
	res := newBitmapContainer()
	if bc, ok := other.(*bitmapContainer); ok {
		for i := 0; i < 1024; i++ {
			res.words[i] = b.words[i] & bc.words[i]
		}
	} else {
		ac := other.(*arrayContainer)
		return ac.and(b)
	}

	if res.cardinality() < arrayThreshold {
		ac := newArrayContainer()
		ac.content = res.toSlice()
		return ac
	}
	return res
}

func (b *bitmapContainer) or(other container) container {
	res := newBitmapContainer()
	if bc, ok := other.(*bitmapContainer); ok {
		for i := 0; i < 1024; i++ {
			res.words[i] = b.words[i] | bc.words[i]
		}
	} else {
		copy(res.words[:], b.words[:])
		for _, v := range other.toSlice() {
			res.add(v)
		}
	}
	return res
}

func (b *bitmapContainer) andNot(other container) container {
	res := newBitmapContainer()
	if bc, ok := other.(*bitmapContainer); ok {
		for i := 0; i < 1024; i++ {
			res.words[i] = b.words[i] &^ bc.words[i]
		}
	} else {
		copy(res.words[:], b.words[:])
		for _, v := range other.toSlice() {
			res.remove(v)
		}
	}

	if res.cardinality() < arrayThreshold {
		ac := newArrayContainer()
		ac.content = res.toSlice()
		return ac
	}
	return res
}
