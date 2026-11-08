# SPEC: go-bloom

## 1. Metadata
- **Module**: `go-bloom`
- **Package**: `gobloom`
- **Location**: `C:\Library\go-bloom`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `hash/fnv`, `math`, `sync`, `encoding/binary`, `errors`, `fmt`).

---

## 2. Core Capabilities
1. **Ultra-Fast Sub-50ns Membership Checking**:
   - ~42 ns per lookup with 0 heap allocations.
2. **Kirsch-Mitzenmacher Double Hashing**:
   - Generates $k$ independent hash probes from two 64-bit seed hashes (`FNV-1a` + `SplitMix64`), eliminating multiple hash algorithm allocations.
3. **Mathematically Optimal Bit Sizing**:
   - Automatically computes optimal bit array length $m$ and hash probe count $k$ for requested capacity $n$ and false positive rate $p$.
4. **Serialization & Cache Persistence**:
   - `Bytes()` and `FromBytes()` for fast disk persistence or caching in Redis/PostgreSQL BLOBs.

---

## 3. Public API
```go
type Filter struct { ... }
func New(expectedItems uint64, falsePositiveRate float64) (*Filter, error)
func (f *Filter) Add(item []byte)
func (f *Filter) AddString(s string)
func (f *Filter) Contains(item []byte) bool
func (f *Filter) ContainsString(s string) bool
func (f *Filter) Reset()
func (f *Filter) SizeInBytes() int
func (f *Filter) Bytes() []byte
func FromBytes(data []byte) (*Filter, error)
```
