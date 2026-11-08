package hashring

import (
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

// HashRing manages a consistent hash ring with virtual nodes for distributed balancing.
type HashRing struct {
	mu       sync.RWMutex
	vnodes   int
	ring     []uint64          // Sorted slice of hash values on the ring
	vnodeMap map[uint64]string // Hash value -> Physical node name
	nodes    map[string]struct{}
}

// New creates a new HashRing with the specified number of virtual nodes per physical node.
// If vnodes <= 0, default 100 virtual nodes are used.
func New(vnodes int, initialNodes ...string) *HashRing {
	if vnodes <= 0 {
		vnodes = 100
	}

	hr := &HashRing{
		vnodes:   vnodes,
		vnodeMap: make(map[uint64]string),
		nodes:    make(map[string]struct{}),
	}

	for _, node := range initialNodes {
		hr.Add(node)
	}

	return hr
}

// Add inserts a physical node into the ring with its assigned virtual nodes.
func (hr *HashRing) Add(node string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.nodes[node]; exists {
		return
	}

	hr.nodes[node] = struct{}{}

	for i := 0; i < hr.vnodes; i++ {
		vnodeKey := node + "#" + strconv.Itoa(i)
		h := hashKey(vnodeKey)
		hr.ring = append(hr.ring, h)
		hr.vnodeMap[h] = node
	}

	sort.Slice(hr.ring, func(i, j int) bool {
		return hr.ring[i] < hr.ring[j]
	})
}

// Remove deletes a physical node and its virtual nodes from the ring.
func (hr *HashRing) Remove(node string) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.nodes[node]; !exists {
		return ErrNodeNotFound
	}

	delete(hr.nodes, node)

	// Rebuild ring without the removed node's vnodes
	newRing := make([]uint64, 0, len(hr.ring)-hr.vnodes)
	for _, h := range hr.ring {
		if hr.vnodeMap[h] == node {
			delete(hr.vnodeMap, h)
		} else {
			newRing = append(newRing, h)
		}
	}
	hr.ring = newRing
	return nil
}

// Get locates the physical node responsible for the given key.
func (hr *HashRing) Get(key string) (string, error) {
	nodes, err := hr.GetN(key, 1)
	if err != nil {
		return "", err
	}
	return nodes[0], nil
}

// GetN returns the top N distinct physical nodes responsible for a key in clockwise order (useful for replicas).
func (hr *HashRing) GetN(key string, n int) ([]string, error) {
	if n <= 0 {
		return nil, ErrInvalidCount
	}

	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.ring) == 0 {
		return nil, ErrEmptyRing
	}

	if n > len(hr.nodes) {
		n = len(hr.nodes)
	}

	h := hashKey(key)
	idx := sort.Search(len(hr.ring), func(i int) bool {
		return hr.ring[i] >= h
	})

	if idx == len(hr.ring) {
		idx = 0 // Wrap around circular ring
	}

	var results []string
	seen := make(map[string]struct{})

	for i := 0; i < len(hr.ring) && len(results) < n; i++ {
		currIdx := (idx + i) % len(hr.ring)
		node := hr.vnodeMap[hr.ring[currIdx]]

		if _, alreadySeen := seen[node]; !alreadySeen {
			seen[node] = struct{}{}
			results = append(results, node)
		}
	}

	return results, nil
}

// Nodes returns a slice of all physical nodes currently in the ring.
func (hr *HashRing) Nodes() []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	res := make([]string, 0, len(hr.nodes))
	for node := range hr.nodes {
		res = append(res, node)
	}
	return res
}

// Len returns the count of physical nodes in the ring.
func (hr *HashRing) Len() int {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return len(hr.nodes)
}

func hashKey(key string) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	h := hasher.Sum64()

	// SplitMix64 finalizer for uniform 64-bit avalanche dispersion
	h ^= h >> 30
	h *= 0xbf58476d1ce4e5b9
	h ^= h >> 27
	h *= 0x94d049bb133111eb
	h ^= h >> 31
	return h
}
