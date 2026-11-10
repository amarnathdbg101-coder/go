package bitmapindex

import (
	"sync"
)

// Version defines the release version of go-bitmapindex.
const Version = "1.0.0"

// Index implements a Columnar Bit-Sliced Index (BSI) for ultra-fast OLAP aggregations.
type Index struct {
	mu        sync.RWMutex
	bitWidth  int
	slices    []*bitset
	existence *bitset
}

// New creates an initialized Bit-Sliced Index with the specified bit width (e.g. 16, 32, 64).
func New(bitWidth int) (*Index, error) {
	if bitWidth <= 0 || bitWidth > 64 {
		bitWidth = 32
	}

	slices := make([]*bitset, bitWidth)
	for i := range slices {
		slices[i] = newBitset()
	}

	return &Index{
		bitWidth:  bitWidth,
		slices:    slices,
		existence: newBitset(),
	}, nil
}

// Set indexes the numeric value for a given row ID.
func (idx *Index) Set(rowID uint32, val uint64) error {
	if idx.bitWidth < 64 && val >= (uint64(1)<<idx.bitWidth) {
		return ErrBitWidthExceeded
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.existence.set(rowID)

	for b := 0; b < idx.bitWidth; b++ {
		if (val & (uint64(1) << b)) != 0 {
			idx.slices[b].set(rowID)
		}
	}
	return nil
}

// Get retrieves the reconstructed integer value for a row ID.
func (idx *Index) Get(rowID uint32) (uint64, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if !idx.existence.test(rowID) {
		return 0, false
	}

	var val uint64
	for b := 0; b < idx.bitWidth; b++ {
		if idx.slices[b].test(rowID) {
			val |= (uint64(1) << b)
		}
	}
	return val, true
}

// FilterRange returns all row IDs whose value falls within [minVal, maxVal].
func (idx *Index) FilterRange(minVal, maxVal uint64) ([]uint32, error) {
	if minVal > maxVal {
		return nil, ErrInvalidRange
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var matches []uint32
	allRows := idx.existence.toSlice()

	for _, row := range allRows {
		// Reconstruct value
		var val uint64
		for b := 0; b < idx.bitWidth; b++ {
			if idx.slices[b].test(row) {
				val |= (uint64(1) << b)
			}
		}
		if val >= minVal && val <= maxVal {
			matches = append(matches, row)
		}
	}
	return matches, nil
}

// Sum computes the exact mathematical sum of values for the given row IDs using hardware POPCNT.
// If matchingRows is nil, computes the sum across the entire column.
func (idx *Index) Sum(matchingRows []uint32) uint64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(matchingRows) == 0 && matchingRows != nil {
		return 0
	}

	var filter *bitset
	if matchingRows != nil {
		filter = newBitset()
		for _, r := range matchingRows {
			filter.set(r)
		}
	}

	var total uint64
	for b := 0; b < idx.bitWidth; b++ {
		multiplier := uint64(1) << b
		var bitCount uint64
		if filter != nil {
			bitCount = idx.slices[b].and(filter).count()
		} else {
			bitCount = idx.slices[b].count()
		}
		total += multiplier * bitCount
	}
	return total
}

// Avg computes the arithmetic mean for the given matching row IDs.
func (idx *Index) Avg(matchingRows []uint32) float64 {
	var count uint64
	if matchingRows != nil {
		count = uint64(len(matchingRows))
	} else {
		idx.mu.RLock()
		count = idx.existence.count()
		idx.mu.RUnlock()
	}

	if count == 0 {
		return 0
	}
	return float64(idx.Sum(matchingRows)) / float64(count)
}

// Count returns the number of non-null indexed rows.
func (idx *Index) Count() uint64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.existence.count()
}
