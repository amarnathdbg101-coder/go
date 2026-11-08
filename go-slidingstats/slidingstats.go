package slidingstats

import (
	"math"
	"sort"
	"sync"
	"time"
)

type statBucket struct {
	timestamp int64 // bucket index
	samples   []float64
}

// Window maintains rolling observations across a ring buffer of time sub-buckets.
type Window struct {
	mu             sync.RWMutex
	windowSize     time.Duration
	bucketDuration time.Duration
	bucketCount    int
	buckets        []statBucket
	maxPerBucket   int
}

// New creates a new sliding window stats accumulator.
func New(windowSize time.Duration, bucketCount int) (*Window, error) {
	if windowSize <= 0 {
		return nil, ErrInvalidWindow
	}
	if bucketCount <= 0 {
		return nil, ErrInvalidBuckets
	}

	bucketDuration := windowSize / time.Duration(bucketCount)
	if bucketDuration <= 0 {
		bucketDuration = time.Millisecond
	}

	return &Window{
		windowSize:     windowSize,
		bucketDuration: bucketDuration,
		bucketCount:    bucketCount,
		buckets:        make([]statBucket, bucketCount),
		maxPerBucket:   500, // bounds RAM to ~500 floats per second
	}, nil
}

func (w *Window) getBucketIndex(now time.Time) (int64, int) {
	currentBucket := now.UnixNano() / int64(w.bucketDuration)
	idx := int(currentBucket % int64(w.bucketCount))
	if idx < 0 {
		idx = -idx
	}
	return currentBucket, idx
}

// Observe records a new numeric sample into the active temporal bucket.
func (w *Window) Observe(val float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	currentBucket, idx := w.getBucketIndex(now)

	b := &w.buckets[idx]
	if b.timestamp != currentBucket {
		// Reset bucket for new time interval
		b.timestamp = currentBucket
		b.samples = b.samples[:0]
	}

	if len(b.samples) < w.maxPerBucket {
		b.samples = append(b.samples, val)
	}
}

// Snapshot captures the current summary statistics and percentiles across the rolling window.
type Snapshot struct {
	Count  int64
	Min    float64
	Max    float64
	Mean   float64
	StdDev float64
	P50    float64
	P90    float64
	P95    float64
	P99    float64
}

// Snapshot computes statistical metrics over all valid samples in the rolling window.
func (w *Window) Snapshot() Snapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	now := time.Now()
	currentBucket := now.UnixNano() / int64(w.bucketDuration)
	oldestValidBucket := currentBucket - int64(w.bucketCount) + 1

	var allSamples []float64
	for i := 0; i < w.bucketCount; i++ {
		b := &w.buckets[i]
		if b.timestamp >= oldestValidBucket && b.timestamp <= currentBucket {
			allSamples = append(allSamples, b.samples...)
		}
	}

	n := len(allSamples)
	if n == 0 {
		return Snapshot{}
	}

	sort.Float64s(allSamples)

	minVal := allSamples[0]
	maxVal := allSamples[n-1]

	var sum float64
	for _, v := range allSamples {
		sum += v
	}
	mean := sum / float64(n)

	var varianceSum float64
	for _, v := range allSamples {
		diff := v - mean
		varianceSum += diff * diff
	}
	stdDev := math.Sqrt(varianceSum / float64(n))

	quantile := func(q float64) float64 {
		if q <= 0 {
			return allSamples[0]
		}
		if q >= 1.0 {
			return allSamples[n-1]
		}
		pos := q * float64(n-1)
		base := int(math.Floor(pos))
		rest := pos - float64(base)
		if base+1 < n {
			return allSamples[base] + rest*(allSamples[base+1]-allSamples[base])
		}
		return allSamples[base]
	}

	return Snapshot{
		Count:  int64(n),
		Min:    minVal,
		Max:    maxVal,
		Mean:   mean,
		StdDev: stdDev,
		P50:    quantile(0.50),
		P90:    quantile(0.90),
		P95:    quantile(0.95),
		P99:    quantile(0.99),
	}
}

// Reset clears all recorded data points.
func (w *Window) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := 0; i < w.bucketCount; i++ {
		w.buckets[i] = statBucket{}
	}
}
