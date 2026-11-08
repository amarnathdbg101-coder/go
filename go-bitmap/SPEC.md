# Technical Specification: go-bitmap

## 1. Domain & Purpose
`go-bitmap` is a compact, high-performance bit array / bitset engine for Go. Storing millions of boolean flags or ID memberships in standard Go maps consumes 8 to 16 bytes per entry; `go-bitmap` compresses this to exactly 1 bit per entry (saving over 98% memory). It supports dynamic capacity expansion, hardware-accelerated population counts (`POPCNT`), and bitwise set operations (`And`, `Or`, `AndNot`).

## 2. Architecture & Design Principles
- **64-Bit Word Chunking**: Uses `[]uint64` slices with bit shift indices (`bit >> 6` and `bit & 0x3F`) for lightning fast $O(1)$ operations.
- **Hardware-Accelerated POPCNT**: Implements `math/bits.OnesCount64` for CPU-level assembly instruction execution when computing set bit counts.
- **Automatic Dynamic Expansion**: Expands slice buffer on-demand if a set bit index exceeds current capacity.
- **Bitwise Set Algebra**: Computes mathematical intersections (`And`), unions (`Or`), and relative complements (`AndNot`) in parallel across 64-bit words.
- **Zero Heavy Dependencies**: Pure Go standard library (`math/bits`, `sync`, `encoding/binary`).

## 3. API Surface
- `type Bitmap struct`
- `func New(initialBits int) *Bitmap`
- `func (b *Bitmap) Set(bit int) error`
- `func (b *Bitmap) Clear(bit int) error`
- `func (b *Bitmap) Toggle(bit int) error`
- `func (b *Bitmap) Test(bit int) bool`
- `func (b *Bitmap) Count() int`
- `func (b *Bitmap) And(other *Bitmap) *Bitmap`
- `func (b *Bitmap) Or(other *Bitmap) *Bitmap`
- `func (b *Bitmap) AndNot(other *Bitmap) *Bitmap`
- `func (b *Bitmap) Reset()`
- `func (b *Bitmap) ToBytes() []byte`
- `func FromBytes(data []byte) *Bitmap`

## 4. Error Handling
- `ErrOutOfBounds`: Negative bit index supplied.
- `ErrSizeMismatch`: Incompatible bitmap operation.

## 5. Verification & Testing
- Unit tests verify word boundary transitions (bit 63 -> 64), dynamic memory expansion past 1000 bits, set operations (And, Or, AndNot), binary byte serialization, and popcount tallying.
