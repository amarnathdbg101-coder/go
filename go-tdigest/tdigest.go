package tdigest

import (
	"math"
	"sort"
	"sync"
)

// Version defines the release version of go-tdigest.
const Version = "1.0.0"

// Centroid represents a clustered group of samples with a mean and total weight.
type Centroid struct {
	Mean   float64
	Weight float64
}

// TDigest implements Ted Dunning's online streaming quantile estimator.
type TDigest struct {
	mu          sync.RWMutex
	compression float64
	centroids   []Centroid
	buffer      []Centroid
	totalWeight float64
	minVal      float64
	maxVal      float64
}

// New creates an initialized t-Digest instance.
func New(opts ...Option) (*TDigest, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Compression <= 0 {
		return nil, ErrInvalidCompression
	}

	return &TDigest{
		compression: cfg.Compression,
		centroids:   make([]Centroid, 0, int(cfg.Compression*2)),
		buffer:      make([]Centroid, 0, int(cfg.Compression*2)),
		minVal:      math.Inf(1),
		maxVal:      math.Inf(-1),
	}, nil
}

// Add inserts a single sample with default weight 1.0.
func (td *TDigest) Add(val float64) {
	td.AddWithWeight(val, 1.0)
}

// AddWithWeight inserts a sample with an associated weight (frequency).
func (td *TDigest) AddWithWeight(val float64, weight float64) {
	if weight <= 0 {
		return
	}

	td.mu.Lock()
	defer td.mu.Unlock()

	if val < td.minVal {
		td.minVal = val
	}
	if val > td.maxVal {
		td.maxVal = val
	}

	td.buffer = append(td.buffer, Centroid{Mean: val, Weight: weight})
	td.totalWeight += weight

	// Compress when buffer capacity is reached
	if len(td.buffer) >= int(td.compression*2) {
		td.compressInternal()
	}
}

// Quantile estimates the value at quantile q in [0.0, 1.0] (e.g. 0.5 for median, 0.99 for P99).
func (td *TDigest) Quantile(q float64) (float64, error) {
	if q < 0.0 || q > 1.0 {
		return 0, ErrInvalidQuantile
	}

	td.mu.Lock()
	defer td.mu.Unlock()

	if td.totalWeight == 0 {
		return 0, ErrEmptyDigest
	}

	if len(td.buffer) > 0 {
		td.compressInternal()
	}

	if len(td.centroids) == 1 {
		return td.centroids[0].Mean, nil
	}

	if q == 0.0 {
		return td.minVal, nil
	}
	if q == 1.0 {
		return td.maxVal, nil
	}

	targetWeight := q * td.totalWeight
	var cumWeight float64

	for i := 0; i < len(td.centroids); i++ {
		cur := td.centroids[i]
		curHalf := cur.Weight / 2.0

		if cumWeight+curHalf >= targetWeight {
			if i == 0 {
				return td.minVal + (targetWeight/curHalf)*(cur.Mean-td.minVal), nil
			}
			prev := td.centroids[i-1]
			prevHalf := prev.Weight / 2.0
			delta := targetWeight - (cumWeight - prevHalf)
			span := prevHalf + curHalf
			frac := delta / span
			return prev.Mean + frac*(cur.Mean-prev.Mean), nil
		}

		cumWeight += cur.Weight
	}

	// Last centroid interpolation to max
	last := td.centroids[len(td.centroids)-1]
	return last.Mean + (td.maxVal-last.Mean)*(q-0.5)/0.5, nil
}

// Merge combines another t-Digest into this digest with full accuracy preservation.
func (td *TDigest) Merge(other *TDigest) error {
	if other == nil {
		return nil
	}

	td.mu.Lock()
	defer td.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	if other.totalWeight == 0 {
		return nil
	}

	if other.minVal < td.minVal {
		td.minVal = other.minVal
	}
	if other.maxVal > td.maxVal {
		td.maxVal = other.maxVal
	}

	td.buffer = append(td.buffer, other.centroids...)
	td.buffer = append(td.buffer, other.buffer...)
	td.totalWeight += other.totalWeight

	td.compressInternal()
	return nil
}

func (td *TDigest) compressInternal() {
	if len(td.buffer) == 0 && len(td.centroids) == 0 {
		return
	}

	all := append(td.centroids, td.buffer...)
	td.buffer = td.buffer[:0]

	sort.Slice(all, func(i, j int) bool {
		return all[i].Mean < all[j].Mean
	})

	var compressed []Centroid
	cur := all[0]
	var weightSoFar float64

	for i := 1; i < len(all); i++ {
		next := all[i]
		proposedWeight := cur.Weight + next.Weight
		q := (weightSoFar + proposedWeight/2.0) / td.totalWeight

		// Scale limit: smaller clusters at extreme tails (q -> 0 and q -> 1)
		maxWeight := 4.0 * td.totalWeight * q * (1.0 - q) / td.compression
		if maxWeight < 1.0 {
			maxWeight = 1.0
		}

		if proposedWeight <= maxWeight {
			// Merge next into cur
			cur.Mean = (cur.Mean*cur.Weight + next.Mean*next.Weight) / proposedWeight
			cur.Weight = proposedWeight
		} else {
			compressed = append(compressed, cur)
			weightSoFar += cur.Weight
			cur = next
		}
	}
	compressed = append(compressed, cur)
	td.centroids = compressed
}

// Count returns the total weight/count of samples added to the digest.
func (td *TDigest) Count() float64 {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.totalWeight
}

// Min returns the minimum observed value.
func (td *TDigest) Min() float64 {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.minVal
}

// Max returns the maximum observed value.
func (td *TDigest) Max() float64 {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.maxVal
}

// Reset clears the digest.
func (td *TDigest) Reset() {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.centroids = td.centroids[:0]
	td.buffer = td.buffer[:0]
	td.totalWeight = 0
	td.minVal = math.Inf(1)
	td.maxVal = math.Inf(-1)
}
