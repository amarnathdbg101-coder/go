# SPEC: go-metrics

## 1. Metadata
- **Module**: `go-metrics`
- **Package**: `gometrics`
- **Location**: `C:\Library\go-metrics`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `sync/atomic`, `runtime`, `net/http`, `math`, `sort`, `strings`, `fmt`, `errors`).

---

## 2. Core Capabilities
1. **Thread-Safe Core Metrics**:
   - `Counter`: Lock-free atomic increment (`CompareAndSwapUint64`).
   - `Gauge`: Lock-free atomic gauge setting, addition, subtraction.
   - `Histogram`: Configurable upper bound buckets, cumulative counts, observation sum, ring-buffered percentile estimations (`Percentile(50.0)`, `Percentile(99.0)`).
2. **Prometheus Exposition Format (v0.0.4)**:
   - Serializes metrics with `# HELP`, `# TYPE`, dimensions `{key="val"}`, `_bucket{le="0.05"}`, `_bucket{le="+Inf"}`, `_sum`, `_count`.
3. **Automatic Go Runtime Collector**:
   - `RegisterRuntimeMetrics(reg)` exposes `go_goroutines`, `go_memstats_alloc_bytes`, `go_memstats_sys_bytes`, `go_memstats_gc_completed_total`.
4. **HTTP Middleware & Scrape Endpoint**:
   - `HTTPMiddleware`: tracks `http_requests_total{method, status}` and `http_request_duration_seconds{method}`.
   - `Handler`: standard `http.HandlerFunc` exposing `/metrics`.

---

## 3. Public API
```go
type Labels map[string]string

type Counter struct { ... }
func (c *Counter) Inc()
func (c *Counter) Add(delta float64)
func (c *Counter) Value() float64

type Gauge struct { ... }
func (g *Gauge) Set(val float64)
func (g *Gauge) Inc()
func (g *Gauge) Dec()
func (g *Gauge) Add(delta float64)
func (g *Gauge) Value() float64

type Histogram struct { ... }
func (h *Histogram) Observe(val float64)
func (h *Histogram) Percentile(p float64) float64
func (h *Histogram) Snapshot() (count uint64, sum float64, buckets []float64, counts []uint64)

type Registry struct { ... }
func NewRegistry() *Registry
func (r *Registry) Counter(name, help string, labels Labels) *Counter
func (r *Registry) Gauge(name, help string, labels Labels) *Gauge
func (r *Registry) Histogram(name, help string, buckets []float64, labels Labels) *Histogram
func (r *Registry) Gather() string

func RegisterRuntimeMetrics(r *Registry) func()
func Handler(r *Registry, samplers ...func()) http.HandlerFunc
func HTTPMiddleware(r *Registry) func(next http.Handler) http.Handler
```
