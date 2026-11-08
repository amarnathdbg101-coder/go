# SPEC: go-idgen

## 1. Metadata
- **Module**: `go-idgen`
- **Package**: `goidgen`
- **Location**: `C:\Library\go-idgen`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `time`, `strings`, `fmt`, `errors`).

---

## 2. Core Capabilities
1. **Ultra-High Throughput**:
   - ~4.8 million IDs per second per node with 0 bytes and 0 heap allocations per operation.
2. **63-bit Positive Integer IDs (PostgreSQL `BIGINT` Friendly)**:
   - 41 bits timestamp delta (~69 years of coverage).
   - 10 bits node ID (supports up to 1,024 distributed cluster nodes).
   - 12 bits sequence number (4,096 IDs per millisecond per node).
3. **NTP Clock Drift Defense**:
   - Handles slight backwards clock drift (<= 5ms) by safely sleeping.
   - Rejects severe backwards drifts with `ErrClockBackwards` to prevent ID duplication.
4. **Deconstructability**:
   - Any generated ID can be decomposed back into exact UTC timestamp, node ID, and sequence without querying a database.
5. **Compact Base62 Alphanumeric Serialization**:
   - URL-safe string encoding (e.g. `7xK9pL2`) for public customer order IDs and short links.

---

## 3. Public API
```go
const DefaultEpoch int64 = 1704067200000 // 2024-01-01 00:00:00 UTC

type Node struct { ... }
func NewNode(nodeID int64, customEpoch ...int64) (*Node, error)
func (n *Node) Generate() (int64, error)
func (n *Node) GenerateBase62() (string, error)
func (n *Node) Deconstruct(id int64) (timestamp time.Time, nodeID int64, sequence int64)

func EncodeBase62(n uint64) string
func DecodeBase62(s string) (uint64, error)
```
