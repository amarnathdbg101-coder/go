# SPEC: go-bytespool

## 1. Metadata
- **Module**: `go-bytespool`
- **Package**: `bytespool`
- **Location**: `C:\Library\go-bytespool`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `strconv`, `io`).

---

## 2. Core Capabilities
1. **Tiered Buffer Architecture**:
   - Routes buffer allocations to one of four isolated `sync.Pool` buckets based on capacity hint:
     - TierSmall: 512 B
     - TierMedium: 4 KB
     - TierLarge: 64 KB
     - TierHuge: 1 MB
2. **Buffer Bloat & Memory Leak Prevention**:
   - Automatically drops buffers that grow past the configurable `maxCap` threshold (default 2 MB) when `Put` is called, preventing resident set size (RSS) creep.
3. **Zero-Allocation Data Appenders**:
   - `WriteString`, `WriteByte`, `WriteInt`, `WriteUint`, and `WriteFloat` avoid intermediate heap allocations by directly formatting into the existing byte slice.
4. **Ergonomic Scoped Safe Helper**:
   - `WithBuffer(hint, fn)` automates `Get` and deferred `Put` recovery.

---

## 3. Public API
```go
const (
    TierSmall  = 512
    TierMedium = 4096
    TierLarge  = 65536
    TierHuge   = 1048576
    DefaultMax = 2097152
)

type Buffer struct
func (b *Buffer) Write(p []byte) (int, error)
func (b *Buffer) WriteString(s string) (int, error)
func (b *Buffer) WriteByte(c byte) error
func (b *Buffer) WriteInt(n int64)
func (b *Buffer) WriteUint(n uint64)
func (b *Buffer) WriteFloat(f float64, prec int)
func (b *Buffer) Bytes() []byte
func (b *Buffer) String() string
func (b *Buffer) Len() int
func (b *Buffer) Cap() int
func (b *Buffer) Reset()
func (b *Buffer) ReadFrom(r io.Reader) (int64, error)

type TieredPool struct
func New(maxCap int) *TieredPool
func (tp *TieredPool) Get(capacityHint int) *Buffer
func (tp *TieredPool) Put(buf *Buffer)

func Get(capacityHint int) *Buffer
func Put(buf *Buffer)
func WithBuffer(capacityHint int, fn func(buf *Buffer))
```
