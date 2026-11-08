package gometrics

import (
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// MetricType identifies the Prometheus metric representation.
type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeHistogram MetricType = "histogram"
)

// Labels represents key-value dimensions for a metric.
type Labels map[string]string

// Format returns Prometheus label string: {k1="v1",k2="v2"}.
func (l Labels) Format() string {
	if len(l) == 0 {
		return ""
	}
	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(k)
		sb.WriteString(`="`)
		sb.WriteString(escapeLabel(l[k]))
		sb.WriteString(`"`)
	}
	sb.WriteString("}")
	return sb.String()
}

func escapeLabel(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, "\n", `\n`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	return v
}

// Counter represents a monotonically increasing value.
type Counter struct {
	valBits uint64 // float64 stored as uint64 bits for atomic ops
}

// NewCounter creates an initialized Counter.
func NewCounter() *Counter {
	return &Counter{}
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.Add(1.0)
}

// Add increases the counter by delta. Ignores negative deltas.
func (c *Counter) Add(delta float64) {
	if delta <= 0 {
		return
	}
	for {
		oldBits := atomic.LoadUint64(&c.valBits)
		oldVal := math.Float64frombits(oldBits)
		newVal := oldVal + delta
		newBits := math.Float64bits(newVal)
		if atomic.CompareAndSwapUint64(&c.valBits, oldBits, newBits) {
			return
		}
	}
}

// Value returns current counter value.
func (c *Counter) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&c.valBits))
}

// Gauge represents a value that can go up and down.
type Gauge struct {
	valBits uint64
}

// NewGauge creates an initialized Gauge.
func NewGauge() *Gauge {
	return &Gauge{}
}

// Set updates the gauge to a specific value.
func (g *Gauge) Set(val float64) {
	atomic.StoreUint64(&g.valBits, math.Float64bits(val))
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() {
	g.Add(1.0)
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() {
	g.Add(-1.0)
}

// Add changes gauge value by delta.
func (g *Gauge) Add(delta float64) {
	for {
		oldBits := atomic.LoadUint64(&g.valBits)
		oldVal := math.Float64frombits(oldBits)
		newVal := oldVal + delta
		newBits := math.Float64bits(newVal)
		if atomic.CompareAndSwapUint64(&g.valBits, oldBits, newBits) {
			return
		}
	}
}

// Value returns current gauge value.
func (g *Gauge) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.valBits))
}

// DefaultBuckets for latency observations (in seconds).
var DefaultBuckets = []float64{
	0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
}

// Histogram tracks sample distributions across configurable buckets.
type Histogram struct {
	mu           sync.RWMutex
	buckets      []float64
	bucketCounts []uint64
	count        uint64
	sum          float64
	samples      []float64 // Retain samples for percentile calculation (ring buffer limit)
	maxSamples   int
}

// NewHistogram creates an initialized Histogram with specified upper bounds.
func NewHistogram(buckets []float64) *Histogram {
	if len(buckets) == 0 {
		buckets = DefaultBuckets
	}
	sortedBuckets := make([]float64, len(buckets))
	copy(sortedBuckets, buckets)
	sort.Float64s(sortedBuckets)

	return &Histogram{
		buckets:      sortedBuckets,
		bucketCounts: make([]uint64, len(sortedBuckets)),
		maxSamples:   2048,
		samples:      make([]float64, 0, 2048),
	}
}

// Observe records a sample value into the histogram.
func (h *Histogram) Observe(val float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.count++
	h.sum += val

	for i, bound := range h.buckets {
		if val <= bound {
			h.bucketCounts[i]++
		}
	}

	if len(h.samples) < h.maxSamples {
		h.samples = append(h.samples, val)
	} else {
		// Overwrite random/earliest slot
		h.samples[h.count%uint64(h.maxSamples)] = val
	}
}

// Snapshot returns count, sum, and bucket counts for export.
func (h *Histogram) Snapshot() (count uint64, sum float64, buckets []float64, counts []uint64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	bCopy := make([]float64, len(h.buckets))
	copy(bCopy, h.buckets)

	cCopy := make([]uint64, len(h.bucketCounts))
	copy(cCopy, h.bucketCounts)

	return h.count, h.sum, bCopy, cCopy
}

// Percentile computes approximate percentile (0.0 to 100.0) from observed samples.
func (h *Histogram) Percentile(p float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.samples) == 0 {
		return 0
	}

	sorted := make([]float64, len(h.samples))
	copy(sorted, h.samples)
	sort.Float64s(sorted)

	rank := (p / 100.0) * float64(len(sorted)-1)
	idx := int(math.Round(rank))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return sorted[idx]
}
