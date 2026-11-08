# SPEC: go-compress

## 1. Metadata
- **Module**: `go-compress`
- **Package**: `gocompress`
- **Location**: `C:\Library\go-compress`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `compress/gzip`, `net/http`, `sync`, `strings`, `bytes`, `io`, `fmt`).

---

## 2. Core Capabilities
1. **Recycled Gzip Pool (`sync.Pool`)**:
   - Reuses `gzip.Writer` allocations to eliminate heap allocations and garbage collection pressure under heavy traffic.
2. **Threshold Buffering**:
   - Only compresses responses exceeding `MinSize` (default 1024 bytes), preventing payload bloat on tiny JSON responses.
3. **Accept-Encoding Negotiation**:
   - Respects client capabilities and WebSocket upgrade requests, maintaining `Vary: Accept-Encoding`.
4. **Standalone Compression**:
   - `Gzip(data)` and `Gunzip(compressed)` for arbitrary byte slices, payloads, or database BLOBs.

---

## 3. Public API
```go
func Gzip(data []byte, level ...int) ([]byte, error)
func Gunzip(compressed []byte) ([]byte, error)

type Config struct {
    MinSize int
    Level   int
}

var DefaultConfig Config

type Compressor struct { ... }
func New(cfg ...Config) *Compressor
func (c *Compressor) Handler(next http.Handler) http.Handler
func Middleware(cfg ...Config) func(http.Handler) http.Handler
```
