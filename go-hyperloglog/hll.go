package hyperloglog

import (
	"hash/fnv"
	"math"
	"math/bits"
	"sync"
)

// Version defines the release version of go-hyperloglog.
const Version = "1.0.0"

// Sketch implements the 64-bit HyperLogLog++ distinct counting algorithm.
type Sketch struct {
	mu        sync.RWMutex
	p         uint8
	m         uint32
	alpha     float64
	registers []uint8
}

// New creates an initialized HyperLogLog sketch.
func New(opts ...Option) (*Sketch, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Precision < 4 || cfg.Precision > 18 {
		return nil, ErrInvalidPrecision
	}

	m := uint32(1) << cfg.Precision
	return &Sketch{
		p:         cfg.Precision,
		m:         m,
		alpha:     computeAlpha(m),
		registers: make([]uint8, m),
	}, nil
}

func computeAlpha(m uint32) float64 {
	switch m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	default:
		return 0.7213 / (1.0 + 1.079/float64(m))
	}
}

// SplitMix64 avalanche mixer to ensure uniform distribution across all 64 bits.
func mix64(z uint64) uint64 {
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// Add inserts raw byte slice into the sketch.
func (s *Sketch) Add(data []byte) {
	h := fnv.New64a()
	_, _ = h.Write(data)
	s.AddHash(mix64(h.Sum64()))
}

// AddString inserts a string into the sketch.
func (s *Sketch) AddString(str string) {
	s.Add([]byte(str))
}

// AddHash inserts a precomputed 64-bit hash.
func (s *Sketch) AddHash(hash uint64) {
	// Uniformly mix input hash
	hash = mix64(hash)

	// Low p bits determine register index
	idx := hash & uint64(s.m-1)

	// High (64 - p) bits determine trailing zeros
	w := hash >> s.p
	sentinel := uint64(1) << (64 - s.p)
	tz := uint8(bits.TrailingZeros64(w|sentinel)) + 1

	s.mu.Lock()
	if tz > s.registers[idx] {
		s.registers[idx] = tz
	}
	s.mu.Unlock()
}

// Estimate calculates the approximate distinct count (cardinality).
func (s *Sketch) Estimate() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sum float64
	zeros := 0

	for _, val := range s.registers {
		sum += math.Pow(2, -float64(val))
		if val == 0 {
			zeros++
		}
	}

	mFloat := float64(s.m)
	rawEstimate := s.alpha * mFloat * mFloat / sum

	// LinearCounting correction for small cardinalities (E <= 2.5 * m)
	if rawEstimate <= 2.5*mFloat {
		if zeros > 0 {
			return uint64(math.Round(mFloat * math.Log(mFloat/float64(zeros))))
		}
	}

	// Large range correction for 64-bit space
	two64 := math.Pow(2, 64)
	if rawEstimate > (two64 / 30.0) {
		return uint64(math.Round(-two64 * math.Log(1.0-rawEstimate/two64)))
	}

	return uint64(math.Round(rawEstimate))
}

// Merge combines another sketch into this sketch (set union). Both must have identical precision.
func (s *Sketch) Merge(other *Sketch) error {
	if other == nil {
		return ErrNilSketch
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	if s.p != other.p {
		return ErrPrecisionMismatch
	}

	for i := uint32(0); i < s.m; i++ {
		if other.registers[i] > s.registers[i] {
			s.registers[i] = other.registers[i]
		}
	}
	return nil
}

// Reset clears all registers.
func (s *Sketch) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.registers {
		s.registers[i] = 0
	}
}

// Clone creates a deep copy of the sketch.
func (s *Sketch) Clone() *Sketch {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp := &Sketch{
		p:         s.p,
		m:         s.m,
		alpha:     s.alpha,
		registers: make([]uint8, s.m),
	}
	copy(cp.registers, s.registers)
	return cp
}

// Precision returns the precision p.
func (s *Sketch) Precision() uint8 {
	return s.p
}
