package crdt

import (
	"encoding/json"
	"sync"
)

// PNCounter represents a state-based Positive-Negative Counter CRDT.
// It supports concurrent increments and decrements across multiple offline replicas
// and converges to the exact same value regardless of message delivery order.
type PNCounter struct {
	mu     sync.RWMutex
	nodeID string
	P      map[string]uint64 `json:"p"`
	N      map[string]uint64 `json:"n"`
}

// NewPNCounter creates a new PN-Counter for the given replica nodeID.
func NewPNCounter(nodeID string) *PNCounter {
	if nodeID == "" {
		nodeID = "node_default"
	}
	return &PNCounter{
		nodeID: nodeID,
		P:      make(map[string]uint64),
		N:      make(map[string]uint64),
	}
}

// Inc increments the counter by delta for this node.
func (c *PNCounter) Inc(delta uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.P[c.nodeID] += delta
}

// Dec decrements the counter by delta for this node.
func (c *PNCounter) Dec(delta uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.N[c.nodeID] += delta
}

// Value computes the effective counter value: sum(P) - sum(N).
func (c *PNCounter) Value() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sumP, sumN uint64
	for _, v := range c.P {
		sumP += v
	}
	for _, v := range c.N {
		sumN += v
	}
	return int64(sumP) - int64(sumN)
}

// NodeID returns the replica ID of this counter.
func (c *PNCounter) NodeID() string {
	return c.nodeID
}

// Merge converges this counter with another counter using the least upper bound (join).
// It takes max(P1[k], P2[k]) and max(N1[k], N2[k]) for every replica key.
// Merge is commutative, associative, and idempotent.
func (c *PNCounter) Merge(other *PNCounter) error {
	if other == nil {
		return ErrNilNode
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	for k, v := range other.P {
		if v > c.P[k] {
			c.P[k] = v
		}
	}
	for k, v := range other.N {
		if v > c.N[k] {
			c.N[k] = v
		}
	}
	return nil
}

type pnCounterDTO struct {
	NodeID string            `json:"node_id"`
	P      map[string]uint64 `json:"p"`
	N      map[string]uint64 `json:"n"`
}

// MarshalJSON serializes the PN-Counter state for network transmission.
func (c *PNCounter) MarshalJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dto := pnCounterDTO{
		NodeID: c.nodeID,
		P:      c.P,
		N:      c.N,
	}
	return json.Marshal(dto)
}

// UnmarshalJSON deserializes a PN-Counter state from network transmission.
func (c *PNCounter) UnmarshalJSON(data []byte) error {
	var dto pnCounterDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodeID = dto.NodeID
	if dto.P != nil {
		c.P = dto.P
	} else {
		c.P = make(map[string]uint64)
	}
	if dto.N != nil {
		c.N = dto.N
	} else {
		c.N = make(map[string]uint64)
	}
	return nil
}
