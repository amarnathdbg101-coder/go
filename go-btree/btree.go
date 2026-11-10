package btree

import (
	"cmp"
	"sort"
	"sync"
)

// Version defines the release version of go-btree.
const Version = "1.0.0"

type node[K cmp.Ordered, V any] struct {
	keys     []K
	values   []V
	children []*node[K, V]
	isLeaf   bool
}

func newNode[K cmp.Ordered, V any](isLeaf bool) *node[K, V] {
	return &node[K, V]{
		keys:     make([]K, 0, 4),
		values:   make([]V, 0, 4),
		children: make([]*node[K, V], 0, 4),
		isLeaf:   isLeaf,
	}
}

// Tree represents a generic, thread-safe B-Tree.
type Tree[K cmp.Ordered, V any] struct {
	mu     sync.RWMutex
	root   *node[K, V]
	degree int
	count  int
}

// New creates a new B-Tree with the specified functional options.
func New[K cmp.Ordered, V any](opts ...Option) (*Tree[K, V], error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Degree < 2 {
		return nil, ErrInvalidDegree
	}

	return &Tree[K, V]{
		root:   newNode[K, V](true),
		degree: cfg.Degree,
	}, nil
}

// Len returns the total number of key-value pairs stored in the tree.
func (t *Tree[K, V]) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.count
}

// Get searches for key and returns its value and whether it was found.
func (t *Tree[K, V]) Get(key K) (V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cur := t.root
	for cur != nil {
		idx := sort.Search(len(cur.keys), func(i int) bool {
			return cur.keys[i] >= key
		})

		if idx < len(cur.keys) && cur.keys[idx] == key {
			return cur.values[idx], true
		}

		if cur.isLeaf {
			break
		}
		cur = cur.children[idx]
	}

	var zero V
	return zero, false
}

// Set inserts or updates a key-value pair in the B-Tree.
func (t *Tree[K, V]) Set(key K, val V) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	maxKeys := 2*t.degree - 1

	// If root is full, split root
	if len(t.root.keys) >= maxKeys {
		oldRoot := t.root
		newR := newNode[K, V](false)
		newR.children = append(newR.children, oldRoot)
		t.splitChild(newR, 0)
		t.root = newR
	}

	return t.insertNonFull(t.root, key, val)
}

func (t *Tree[K, V]) splitChild(parent *node[K, V], i int) {
	deg := t.degree
	child := parent.children[i]
	sibling := newNode[K, V](child.isLeaf)

	// Middle element
	midIdx := deg - 1
	midKey := child.keys[midIdx]
	midVal := child.values[midIdx]

	// Sibling gets keys from midIdx+1 onwards
	sibling.keys = append(sibling.keys, child.keys[midIdx+1:]...)
	sibling.values = append(sibling.values, child.values[midIdx+1:]...)

	if !child.isLeaf {
		sibling.children = append(sibling.children, child.children[midIdx+1:]...)
		child.children = child.children[:midIdx+1]
	}

	child.keys = child.keys[:midIdx]
	child.values = child.values[:midIdx]

	// Insert midKey into parent at index i
	parent.keys = append(parent.keys, midKey)
	parent.values = append(parent.values, midVal)
	copy(parent.keys[i+1:], parent.keys[i:])
	copy(parent.values[i+1:], parent.values[i:])
	parent.keys[i] = midKey
	parent.values[i] = midVal

	// Insert sibling pointer into parent children at index i+1
	parent.children = append(parent.children, sibling)
	copy(parent.children[i+2:], parent.children[i+1:])
	parent.children[i+1] = sibling
}

func (t *Tree[K, V]) insertNonFull(n *node[K, V], key K, val V) (V, bool) {
	idx := sort.Search(len(n.keys), func(i int) bool {
		return n.keys[i] >= key
	})

	// Exact match: update existing value
	if idx < len(n.keys) && n.keys[idx] == key {
		old := n.values[idx]
		n.values[idx] = val
		return old, true
	}

	if n.isLeaf {
		// Insert into leaf
		n.keys = append(n.keys, key)
		n.values = append(n.values, val)
		copy(n.keys[idx+1:], n.keys[idx:])
		copy(n.values[idx+1:], n.values[idx:])
		n.keys[idx] = key
		n.values[idx] = val
		t.count++
		var zero V
		return zero, false
	}

	// Internal node: check if child is full before descending
	maxKeys := 2*t.degree - 1
	if len(n.children[idx].keys) >= maxKeys {
		t.splitChild(n, idx)
		if key > n.keys[idx] {
			idx++
		} else if key == n.keys[idx] {
			old := n.values[idx]
			n.values[idx] = val
			return old, true
		}
	}

	return t.insertNonFull(n.children[idx], key, val)
}

// Delete removes a key from the tree. Returns previous value and whether it was found.
func (t *Tree[K, V]) Delete(key K) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	val, found := t.deleteInternal(t.root, key)
	if found {
		t.count--
		// If root became empty and has child, promote child
		if len(t.root.keys) == 0 && !t.root.isLeaf {
			t.root = t.root.children[0]
		}
	}
	return val, found
}

func (t *Tree[K, V]) deleteInternal(n *node[K, V], key K) (V, bool) {
	idx := sort.Search(len(n.keys), func(i int) bool {
		return n.keys[i] >= key
	})

	if idx < len(n.keys) && n.keys[idx] == key {
		val := n.values[idx]
		if n.isLeaf {
			// Remove from leaf
			n.keys = append(n.keys[:idx], n.keys[idx+1:]...)
			n.values = append(n.values[:idx], n.values[idx+1:]...)
			return val, true
		}

		// Replace with predecessor from left child
		predNode := n.children[idx]
		for !predNode.isLeaf {
			predNode = predNode.children[len(predNode.children)-1]
		}
		predKey := predNode.keys[len(predNode.keys)-1]
		predVal := predNode.values[len(predNode.values)-1]

		n.keys[idx] = predKey
		n.values[idx] = predVal

		_, _ = t.deleteInternal(n.children[idx], predKey)
		return val, true
	}

	if n.isLeaf {
		var zero V
		return zero, false
	}

	return t.deleteInternal(n.children[idx], key)
}

// Min returns the minimum key-value pair in the tree.
func (t *Tree[K, V]) Min() (K, V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.count == 0 {
		var zk K
		var zv V
		return zk, zv, false
	}

	cur := t.root
	for !cur.isLeaf {
		cur = cur.children[0]
	}
	return cur.keys[0], cur.values[0], true
}

// Max returns the maximum key-value pair in the tree.
func (t *Tree[K, V]) Max() (K, V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.count == 0 {
		var zk K
		var zv V
		return zk, zv, false
	}

	cur := t.root
	for !cur.isLeaf {
		cur = cur.children[len(cur.children)-1]
	}
	lastIdx := len(cur.keys) - 1
	return cur.keys[lastIdx], cur.values[lastIdx], true
}

// Ascend iterates over all key-value pairs in ascending order. If fn returns false, iteration stops.
func (t *Tree[K, V]) Ascend(fn func(key K, value V) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.ascendNode(t.root, fn)
}

func (t *Tree[K, V]) ascendNode(n *node[K, V], fn func(K, V) bool) bool {
	for i := 0; i < len(n.keys); i++ {
		if !n.isLeaf {
			if !t.ascendNode(n.children[i], fn) {
				return false
			}
		}
		if !fn(n.keys[i], n.values[i]) {
			return false
		}
	}
	if !n.isLeaf {
		return t.ascendNode(n.children[len(n.keys)], fn)
	}
	return true
}

// AscendRange iterates over keys in range [start, end] in ascending order.
func (t *Tree[K, V]) AscendRange(start, end K, fn func(key K, value V) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.ascendRangeNode(t.root, start, end, fn)
}

func (t *Tree[K, V]) ascendRangeNode(n *node[K, V], start, end K, fn func(K, V) bool) bool {
	for i := 0; i < len(n.keys); i++ {
		if !n.isLeaf && n.keys[i] >= start {
			if !t.ascendRangeNode(n.children[i], start, end, fn) {
				return false
			}
		}
		if n.keys[i] >= start && n.keys[i] <= end {
			if !fn(n.keys[i], n.values[i]) {
				return false
			}
		}
		if n.keys[i] > end {
			return false
		}
	}
	if !n.isLeaf {
		return t.ascendRangeNode(n.children[len(n.keys)], start, end, fn)
	}
	return true
}
