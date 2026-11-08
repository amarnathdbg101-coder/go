# SPEC: go-slidingstats

## 1. Metadata
- **Module**: `go-slidingstats`
- **Package**: `slidingstats`
- **Location**: `C:\Library\go-slidingstats`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `math`, `sort`, `sync`, `time`, `errors`).

---

## 2. Core Capabilities
1. **Rolling Time Window Histogram**:
   - Records numeric observations across a ring buffer of granular time sub-buckets.
   - Older sub-buckets expire automatically without background timer goroutines.
2. **Comprehensive Statistical Metrics**:
   - Computes `Count`, `Min`, `Max`, `Mean`, `StdDev`, `P50`, `P90`, `P95`, and `P99` percentiles on demand.
3. **RAM-Bounded Reservoir Sampling**:
   - Bounds memory usage by capping the maximum samples collected per sub-bucket (`maxPerBucket`).
4. **Thread-Safe & Lock-Free Multi-Threaded Observations**:
   - Synchronized via `sync.RWMutex`.

---

## 3. Public API
```go
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

type Window struct
func New(windowSize time.Duration, bucketCount int) (*Window, error)
func (w *Window) Observe(val float64)
func (w *Window) Snapshot() Snapshot
func (w *Window) Reset()
```
