package rbtree

import (
	"cmp"
	"sync"
)

// Version defines the release version of go-rbtree.
const Version = "1.0.0"

const (
	red   = true
	black = false
)

type node[K cmp.Ordered, V any] struct {
	key   K
	val   V
	color bool
	left  *node[K, V]
	right *node[K, V]
	size  int
}

func newNode[K cmp.Ordered, V any](key K, val V, color bool) *node[K, V] {
	return &node[K, V]{
		key:   key,
		val:   val,
		color: color,
		size:  1,
	}
}

// Tree represents a generic, thread-safe Left-Leaning Red-Black (LLRB) balanced binary search tree.
type Tree[K cmp.Ordered, V any] struct {
	mu   sync.RWMutex
	root *node[K, V]
}

// New creates an initialized Left-Leaning Red-Black tree.
func New[K cmp.Ordered, V any]() *Tree[K, V] {
	return &Tree[K, V]{}
}

// Len returns the number of key-value elements in the tree.
func (t *Tree[K, V]) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return size(t.root)
}

func size[K cmp.Ordered, V any](n *node[K, V]) int {
	if n == nil {
		return 0
	}
	return n.size
}

func isRed[K cmp.Ordered, V any](n *node[K, V]) bool {
	if n == nil {
		return false
	}
	return n.color == red
}

// Insert inserts or updates a key-value pair in the tree.
func (t *Tree[K, V]) Insert(key K, val V) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var oldVal V
	var updated bool

	t.root, oldVal, updated = t.insertInternal(t.root, key, val)
	t.root.color = black
	return oldVal, updated
}

func (t *Tree[K, V]) insertInternal(h *node[K, V], key K, val V) (*node[K, V], V, bool) {
	if h == nil {
		return newNode(key, val, red), val, false
	}

	var old V
	var updated bool

	if key < h.key {
		h.left, old, updated = t.insertInternal(h.left, key, val)
	} else if key > h.key {
		h.right, old, updated = t.insertInternal(h.right, key, val)
	} else {
		old = h.val
		h.val = val
		return h, old, true
	}

	// Fix right-leaning red links
	if isRed(h.right) && !isRed(h.left) {
		h = rotateLeft(h)
	}
	// Fix two consecutive red links on left
	if isRed(h.left) && isRed(h.left.left) {
		h = rotateRight(h)
	}
	// Split 4-nodes
	if isRed(h.left) && isRed(h.right) {
		flipColors(h)
	}

	h.size = size(h.left) + size(h.right) + 1
	return h, old, updated
}

func rotateLeft[K cmp.Ordered, V any](h *node[K, V]) *node[K, V] {
	x := h.right
	h.right = x.left
	x.left = h
	x.color = h.color
	h.color = red
	x.size = h.size
	h.size = size(h.left) + size(h.right) + 1
	return x
}

func rotateRight[K cmp.Ordered, V any](h *node[K, V]) *node[K, V] {
	x := h.left
	h.left = x.right
	x.right = h
	x.color = h.color
	h.color = red
	x.size = h.size
	h.size = size(h.left) + size(h.right) + 1
	return x
}

func flipColors[K cmp.Ordered, V any](h *node[K, V]) {
	h.color = !h.color
	if h.left != nil {
		h.left.color = !h.left.color
	}
	if h.right != nil {
		h.right.color = !h.right.color
	}
}

// Get searches for the key and returns its value.
func (t *Tree[K, V]) Get(key K) (V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cur := t.root
	for cur != nil {
		if key < cur.key {
			cur = cur.left
		} else if key > cur.key {
			cur = cur.right
		} else {
			return cur.val, true
		}
	}
	var zero V
	return zero, false
}

// Min returns the smallest key in the tree.
func (t *Tree[K, V]) Min() (K, V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.root == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	cur := t.root
	for cur.left != nil {
		cur = cur.left
	}
	return cur.key, cur.val, true
}

// Max returns the largest key in the tree.
func (t *Tree[K, V]) Max() (K, V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.root == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	cur := t.root
	for cur.right != nil {
		cur = cur.right
	}
	return cur.key, cur.val, true
}

// Floor finds the largest key less than or equal to the query key.
func (t *Tree[K, V]) Floor(key K) (K, V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	res := t.floorInternal(t.root, key)
	if res == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	return res.key, res.val, true
}

func (t *Tree[K, V]) floorInternal(h *node[K, V], key K) *node[K, V] {
	if h == nil {
		return nil
	}
	if key == h.key {
		return h
	}
	if key < h.key {
		return t.floorInternal(h.left, key)
	}
	tRight := t.floorInternal(h.right, key)
	if tRight != nil {
		return tRight
	}
	return h
}

// Ceiling finds the smallest key greater than or equal to the query key.
func (t *Tree[K, V]) Ceiling(key K) (K, V, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	res := t.ceilingInternal(t.root, key)
	if res == nil {
		var zk K
		var zv V
		return zk, zv, false
	}
	return res.key, res.val, true
}

func (t *Tree[K, V]) ceilingInternal(h *node[K, V], key K) *node[K, V] {
	if h == nil {
		return nil
	}
	if key == h.key {
		return h
	}
	if key > h.key {
		return t.ceilingInternal(h.right, key)
	}
	tLeft := t.ceilingInternal(h.left, key)
	if tLeft != nil {
		return tLeft
	}
	return h
}

// Ascend iterates over all key-value pairs in sorted ascending order.
func (t *Tree[K, V]) Ascend(fn func(key K, val V) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.ascendInternal(t.root, fn)
}

func (t *Tree[K, V]) ascendInternal(h *node[K, V], fn func(K, V) bool) bool {
	if h == nil {
		return true
	}
	if !t.ascendInternal(h.left, fn) {
		return false
	}
	if !fn(h.key, h.val) {
		return false
	}
	return t.ascendInternal(h.right, fn)
}

// AscendRange iterates over key-value pairs in range [start, end].
func (t *Tree[K, V]) AscendRange(start, end K, fn func(key K, val V) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.ascendRangeInternal(t.root, start, end, fn)
}

func (t *Tree[K, V]) ascendRangeInternal(h *node[K, V], start, end K, fn func(K, V) bool) bool {
	if h == nil {
		return true
	}
	if h.key >= start {
		if !t.ascendRangeInternal(h.left, start, end, fn) {
			return false
		}
	}
	if h.key >= start && h.key <= end {
		if !fn(h.key, h.val) {
			return false
		}
	}
	if h.key <= end {
		return t.ascendRangeInternal(h.right, start, end, fn)
	}
	return true
}
