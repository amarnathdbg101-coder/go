package keyrotator

import "sync/atomic"

// RoundRobinStrategy provides atomic round-robin key selection across healthy keys.
type RoundRobinStrategy struct {
	counter uint64
}

// NewRoundRobinStrategy initializes a round-robin strategy.
func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{}
}

// Select picks the next healthy key in cyclic sequence.
func (s *RoundRobinStrategy) Select(healthy []*KeyHealth) (*KeyHealth, error) {
	n := len(healthy)
	if n == 0 {
		return nil, ErrAllKeysExhausted
	}

	idx := atomic.AddUint64(&s.counter, 1) - 1
	return healthy[idx%uint64(n)], nil
}
