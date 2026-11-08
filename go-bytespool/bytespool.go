package bytespool

import (
	"io"
	"strconv"
	"sync"
)

const (
	TierSmall  = 512              // 512 B
	TierMedium = 4 * 1024         // 4 KB
	TierLarge  = 64 * 1024        // 64 KB
	TierHuge   = 1024 * 1024      // 1 MB
	DefaultMax = 2 * 1024 * 1024  // 2 MB max cap threshold
)

// Buffer is a high-performance growable byte slice with zero-allocation utility writers.
type Buffer struct {
	buf  []byte
	tier int
}

// Write appends bytes to the buffer and implements io.Writer.
func (b *Buffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteString appends a string to the buffer without heap allocation.
func (b *Buffer) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// WriteByte appends a single byte.
func (b *Buffer) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteInt appends an integer formatted as decimal text.
func (b *Buffer) WriteInt(n int64) {
	b.buf = strconv.AppendInt(b.buf, n, 10)
}

// WriteUint appends an unsigned integer formatted as decimal text.
func (b *Buffer) WriteUint(n uint64) {
	b.buf = strconv.AppendUint(b.buf, n, 10)
}

// WriteFloat appends a float formatted with given precision.
func (b *Buffer) WriteFloat(f float64, prec int) {
	b.buf = strconv.AppendFloat(b.buf, f, 'f', prec, 64)
}

// Bytes returns the slice of accumulated bytes.
func (b *Buffer) Bytes() []byte {
	return b.buf
}

// String returns a string copy of the buffer content.
func (b *Buffer) String() string {
	return string(b.buf)
}

// Len returns the current number of bytes in the buffer.
func (b *Buffer) Len() int {
	return len(b.buf)
}

// Cap returns the allocated capacity.
func (b *Buffer) Cap() int {
	return cap(b.buf)
}

// Reset clears the buffer length while preserving underlying capacity.
func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
}

// ReadFrom reads data from r until EOF and appends it to the buffer.
func (b *Buffer) ReadFrom(r io.Reader) (int64, error) {
	var total int64
	for {
		if len(b.buf) == cap(b.buf) {
			// Grow buffer
			b.buf = append(b.buf, 0)[:len(b.buf)]
		}
		free := b.buf[len(b.buf):cap(b.buf)]
		n, err := r.Read(free)
		if n > 0 {
			b.buf = b.buf[:len(b.buf)+n]
			total += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return total, err
		}
	}
}

// TieredPool manages multiple capacity-sized sync.Pool instances with leak prevention.
type TieredPool struct {
	poolSmall  sync.Pool
	poolMedium sync.Pool
	poolLarge  sync.Pool
	poolHuge   sync.Pool
	maxCap     int
}

// New creates a new TieredPool with custom max capacity discard threshold.
func New(maxCap int) *TieredPool {
	if maxCap <= 0 {
		maxCap = DefaultMax
	}
	tp := &TieredPool{
		maxCap: maxCap,
	}

	tp.poolSmall.New = func() any {
		return &Buffer{buf: make([]byte, 0, TierSmall), tier: TierSmall}
	}
	tp.poolMedium.New = func() any {
		return &Buffer{buf: make([]byte, 0, TierMedium), tier: TierMedium}
	}
	tp.poolLarge.New = func() any {
		return &Buffer{buf: make([]byte, 0, TierLarge), tier: TierLarge}
	}
	tp.poolHuge.New = func() any {
		return &Buffer{buf: make([]byte, 0, TierHuge), tier: TierHuge}
	}

	return tp
}

// Default is the package-level singleton TieredPool.
var Default = New(DefaultMax)

// Get routes to the optimal buffer tier based on capacityHint.
func (tp *TieredPool) Get(capacityHint int) *Buffer {
	var buf *Buffer

	switch {
	case capacityHint <= TierSmall:
		buf = tp.poolSmall.Get().(*Buffer)
	case capacityHint <= TierMedium:
		buf = tp.poolMedium.Get().(*Buffer)
	case capacityHint <= TierLarge:
		buf = tp.poolLarge.Get().(*Buffer)
	default:
		buf = tp.poolHuge.Get().(*Buffer)
	}

	buf.Reset()
	return buf
}

// Put resets and returns the buffer to the appropriate tier pool.
// If the buffer grew beyond maxCap, it is discarded to avoid memory bloat.
func (tp *TieredPool) Put(buf *Buffer) {
	if buf == nil {
		return
	}

	// Discard bloat
	if cap(buf.buf) > tp.maxCap {
		return
	}

	buf.Reset()

	switch buf.tier {
	case TierSmall:
		tp.poolSmall.Put(buf)
	case TierMedium:
		tp.poolMedium.Put(buf)
	case TierLarge:
		tp.poolLarge.Put(buf)
	case TierHuge:
		tp.poolHuge.Put(buf)
	default:
		// Not recognized tier, allow GC
	}
}

// Package-level helpers forwarding to Default:

// Get acquires a buffer from Default tiered pool.
func Get(capacityHint int) *Buffer {
	return Default.Get(capacityHint)
}

// Put returns a buffer to Default tiered pool.
func Put(buf *Buffer) {
	Default.Put(buf)
}

// WithBuffer acquires a buffer, passes it to fn, and ensures it is safely Put back via defer.
func WithBuffer(capacityHint int, fn func(buf *Buffer)) {
	b := Get(capacityHint)
	defer Put(b)
	fn(b)
}
