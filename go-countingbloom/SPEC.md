# SPEC: go-countingbloom

## 1. Metadata
- **Module**: `go-countingbloom`
- **Package**: `countingbloom`
- **Location**: `C:\Library\go-countingbloom`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `hash/fnv`, `math`, `sync`, `errors`).

---

## 2. Core Capabilities
1. **Deletable Bloom Filter Engine**:
   - Uses 8-bit counters (`uint8`) instead of single bits, allowing safe removal of keys (`Remove`) without rebuilding the entire filter.
2. **Double-Hashing Optimization**:
   - Kirsch-Mitzenmacher double-hashing (FNV-1a + SplitMix64) generates $K$ independent bucket positions in $O(1)$ time.
3. **Counter Saturation Protection**:
   - Counters max out at 255 rather than wrapping around to 0, ensuring false negative probability remains strictly 0%.
4. **Thread-Safe Concurrent Mode**:
   - `NewConcurrent` wraps operations in `sync.RWMutex`.

---

## 3. Public API
```go
type Filter struct
func New(expectedElements uint, falsePositiveRate float64) (*Filter, error)
func NewConcurrent(expectedElements uint, falsePositiveRate float64) (*Filter, error)
func (f *Filter) Add(data []byte)
func (f *Filter) AddString(s string)
func (f *Filter) Contains(data []byte) bool
func (f *Filter) ContainsString(s string) bool
func (f *Filter) Remove(data []byte) bool
func (f *Filter) RemoveString(s string) bool
func (f *Filter) Count() uint64
func (f *Filter) Clear()
```
