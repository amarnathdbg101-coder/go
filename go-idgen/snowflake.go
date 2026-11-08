package goidgen

import (
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultEpoch: 2024-01-01 00:00:00 UTC in milliseconds
	DefaultEpoch int64 = 1704067200000

	nodeBits     = 10
	sequenceBits = 12

	maxNode     = -1 ^ (-1 << nodeBits)     // 1023
	maxSequence = -1 ^ (-1 << sequenceBits) // 4095

	timeShift = nodeBits + sequenceBits // 22
	nodeShift = sequenceBits            // 12
)

// Node generates 64-bit globally unique, time-ordered Snowflake IDs.
type Node struct {
	mu       sync.Mutex
	epoch    int64
	nodeID   int64
	step     int64
	lastTime int64
}

// NewNode constructs a Snowflake generator for the specified node ID (0-1023).
func NewNode(nodeID int64, customEpoch ...int64) (*Node, error) {
	if nodeID < 0 || nodeID > maxNode {
		return nil, fmt.Errorf("%w: got %d (max %d)", ErrInvalidNodeID, nodeID, maxNode)
	}

	epoch := DefaultEpoch
	if len(customEpoch) > 0 && customEpoch[0] > 0 {
		epoch = customEpoch[0]
	}

	return &Node{
		epoch:    epoch,
		nodeID:   nodeID,
		step:     0,
		lastTime: 0,
	}, nil
}

// Generate creates a unique 64-bit integer ID.
// Safe for concurrent access across multiple goroutines.
func (n *Node) Generate() (int64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()

	if now < n.lastTime {
		drift := n.lastTime - now
		if drift <= 5 {
			// Small NTP clock drift: wait until time catches up
			time.Sleep(time.Duration(drift+1) * time.Millisecond)
			now = time.Now().UnixMilli()
		} else {
			return 0, fmt.Errorf("%w: clock moved backwards by %d ms", ErrClockBackwards, drift)
		}
	}

	if now == n.lastTime {
		n.step = (n.step + 1) & maxSequence
		if n.step == 0 {
			// Sequence exhausted in this millisecond; wait for next millisecond
			for now <= n.lastTime {
				time.Sleep(50 * time.Microsecond)
				now = time.Now().UnixMilli()
			}
		}
	} else {
		n.step = 0
	}

	n.lastTime = now

	// 63-bit integer: [41 bits time delta | 10 bits node ID | 12 bits sequence]
	id := ((now - n.epoch) << timeShift) |
		(n.nodeID << nodeShift) |
		n.step

	return id, nil
}

// GenerateBase62 generates a unique ID and encodes it into a compact alphanumeric Base62 string.
func (n *Node) GenerateBase62() (string, error) {
	id, err := n.Generate()
	if err != nil {
		return "", err
	}
	return EncodeBase62(uint64(id)), nil
}

// Deconstruct decomposes a Snowflake ID into its timestamp, node ID, and sequence number.
func (n *Node) Deconstruct(id int64) (timestamp time.Time, nodeID int64, sequence int64) {
	timeDelta := id >> timeShift
	nodeID = (id >> nodeShift) & maxNode
	sequence = id & maxSequence
	milli := timeDelta + n.epoch
	timestamp = time.UnixMilli(milli).UTC()
	return timestamp, nodeID, sequence
}
