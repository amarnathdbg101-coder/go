package simhash

import (
	"hash/fnv"
	"math/bits"
	"sync"
)

// Version defines the release version of go-simhash.
const Version = "1.0.0"

// Fingerprint represents a 64-bit locality sensitive hash.
type Fingerprint uint64

// Compute computes a 64-bit SimHash fingerprint from a slice of weighted features.
func Compute(features []Feature) Fingerprint {
	if len(features) == 0 {
		return 0
	}

	var v [64]int

	for _, feat := range features {
		weight := feat.Weight
		if weight <= 0 {
			weight = 1
		}

		// 64-bit FNV-1a hash of token
		h := fnv.New64a()
		_, _ = h.Write([]byte(feat.Token))
		hash := h.Sum64()

		for i := 0; i < 64; i++ {
			bit := (hash >> i) & 1
			if bit == 1 {
				v[i] += weight
			} else {
				v[i] -= weight
			}
		}
	}

	var result uint64
	for i := 0; i < 64; i++ {
		if v[i] > 0 {
			result |= (1 << i)
		}
	}

	return Fingerprint(result)
}

// ComputeFromText is a convenience function that tokenizes and shingles text into 2-grams before computing SimHash.
func ComputeFromText(text string) Fingerprint {
	features := FeaturesFromText(text, 2)
	return Compute(features)
}

// Distance computes the Hamming distance (number of differing bits) between two fingerprints.
func (f Fingerprint) Distance(other Fingerprint) int {
	return bits.OnesCount64(uint64(f ^ other))
}

// Similarity returns a normalized similarity score between 0.0 (completely dissimilar) and 1.0 (identical).
func (f Fingerprint) Similarity(other Fingerprint) float64 {
	dist := f.Distance(other)
	return 1.0 - (float64(dist) / 64.0)
}

// -------------------------------------------------------------
// In-Memory Near-Duplicate Search Index
// -------------------------------------------------------------

// Match represents a search result with its calculated distance and similarity.
type Match struct {
	ID         string
	Distance   int
	Similarity float64
}

// Index stores documents and performs fast Hamming distance nearest neighbor searches.
type Index struct {
	mu   sync.RWMutex
	docs map[string]Fingerprint
}

// NewIndex creates an empty SimHash index.
func NewIndex() *Index {
	return &Index{
		docs: make(map[string]Fingerprint),
	}
}

// Insert adds or updates a document fingerprint in the index.
func (idx *Index) Insert(id string, fp Fingerprint) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.docs[id] = fp
}

// Remove deletes a document from the index.
func (idx *Index) Remove(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.docs, id)
}

// Count returns the total number of documents in the index.
func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.docs)
}

// FindNearDuplicates returns all indexed documents whose Hamming distance to target is <= maxDistance.
func (idx *Index) FindNearDuplicates(target Fingerprint, maxDistance int) []Match {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var matches []Match
	for id, fp := range idx.docs {
		dist := target.Distance(fp)
		if dist <= maxDistance {
			matches = append(matches, Match{
				ID:         id,
				Distance:   dist,
				Similarity: 1.0 - (float64(dist) / 64.0),
			})
		}
	}
	return matches
}
