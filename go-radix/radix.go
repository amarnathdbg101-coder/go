package radix

import (
	"sort"
	"strings"
	"sync"
)

// Version defines the release version of go-radix.
const Version = "1.0.0"

type edge[T any] struct {
	label string
	node  *node[T]
}

type node[T any] struct {
	hasValue bool
	val      T
	edges    []edge[T]
}

func newNode[T any]() *node[T] {
	return &node[T]{
		edges: make([]edge[T], 0, 2),
	}
}

// Tree represents a thread-safe, generic Compact Radix Tree (Patricia Trie).
type Tree[T any] struct {
	mu   sync.RWMutex
	root *node[T]
	size int
}

// New creates an empty Radix Tree.
func New[T any]() *Tree[T] {
	return &Tree[T]{
		root: newNode[T](),
	}
}

// Len returns the number of elements in the tree.
func (t *Tree[T]) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

// Insert adds or updates a key in the Radix Tree with path compression and edge splitting.
func (t *Tree[T]) Insert(key string, val T) (T, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	cur := t.root
	search := key

	for {
		if len(search) == 0 {
			var old T
			updated := cur.hasValue
			if updated {
				old = cur.val
			} else {
				t.size++
			}
			cur.hasValue = true
			cur.val = val
			return old, updated
		}

		// Find matching edge by first byte
		idx := -1
		for i := range cur.edges {
			if cur.edges[i].label[0] == search[0] {
				idx = i
				break
			}
		}

		if idx == -1 {
			// No matching edge: create new leaf edge
			leaf := newNode[T]()
			leaf.hasValue = true
			leaf.val = val
			cur.edges = append(cur.edges, edge[T]{
				label: search,
				node:  leaf,
			})
			t.size++
			var zero T
			return zero, false
		}

		// Found edge with common prefix
		matchedEdge := &cur.edges[idx]
		commonLen := commonPrefixLen(search, matchedEdge.label)

		if commonLen == len(matchedEdge.label) {
			// Entire edge label matched: descend into child
			search = search[commonLen:]
			cur = matchedEdge.node
			continue
		}

		// Split edge
		splitNode := newNode[T]()
		oldChild := matchedEdge.node
		remainingEdgeLabel := matchedEdge.label[commonLen:]

		// Split node adopts remaining portion of previous edge
		splitNode.edges = append(splitNode.edges, edge[T]{
			label: remainingEdgeLabel,
			node:  oldChild,
		})

		// Current edge label is shortened to common prefix
		matchedEdge.label = search[:commonLen]
		matchedEdge.node = splitNode

		if commonLen == len(search) {
			// Key terminates at split node
			splitNode.hasValue = true
			splitNode.val = val
			t.size++
			var zero T
			return zero, false
		}

		// Add new leaf for remaining search key
		leaf := newNode[T]()
		leaf.hasValue = true
		leaf.val = val
		splitNode.edges = append(splitNode.edges, edge[T]{
			label: search[commonLen:],
			node:  leaf,
		})
		t.size++
		var zero T
		return zero, false
	}
}

// Get searches for an exact key match.
func (t *Tree[T]) Get(key string) (T, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cur := t.root
	search := key

	for {
		if len(search) == 0 {
			if cur.hasValue {
				return cur.val, true
			}
			var zero T
			return zero, false
		}

		idx := -1
		for i := range cur.edges {
			if cur.edges[i].label[0] == search[0] {
				idx = i
				break
			}
		}
		if idx == -1 {
			var zero T
			return zero, false
		}

		e := cur.edges[idx]
		if !strings.HasPrefix(search, e.label) {
			var zero T
			return zero, false
		}

		search = search[len(e.label):]
		cur = e.node
	}
}

// LongestPrefix finds the longest registered prefix that matches the query string.
func (t *Tree[T]) LongestPrefix(query string) (string, T, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var lastPrefix string
	var lastVal T
	var found bool

	cur := t.root
	search := query

	for {
		if cur.hasValue {
			lastPrefix = query[:len(query)-len(search)]
			lastVal = cur.val
			found = true
		}

		if len(search) == 0 {
			break
		}

		idx := -1
		for i := range cur.edges {
			if cur.edges[i].label[0] == search[0] {
				idx = i
				break
			}
		}
		if idx == -1 {
			break
		}

		e := cur.edges[idx]
		if !strings.HasPrefix(search, e.label) {
			break
		}

		search = search[len(e.label):]
		cur = e.node
	}

	return lastPrefix, lastVal, found
}

// Delete removes a key from the Radix Tree.
func (t *Tree[T]) Delete(key string) (T, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	val, deleted := t.deleteInternal(t.root, key)
	if deleted {
		t.size--
	}
	return val, deleted
}

func (t *Tree[T]) deleteInternal(cur *node[T], search string) (T, bool) {
	if len(search) == 0 {
		if !cur.hasValue {
			var zero T
			return zero, false
		}
		val := cur.val
		cur.hasValue = false
		var zero T
		cur.val = zero
		return val, true
	}

	idx := -1
	for i := range cur.edges {
		if cur.edges[i].label[0] == search[0] {
			idx = i
			break
		}
	}
	if idx == -1 {
		var zero T
		return zero, false
	}

	e := &cur.edges[idx]
	if !strings.HasPrefix(search, e.label) {
		var zero T
		return zero, false
	}

	val, deleted := t.deleteInternal(e.node, search[len(e.label):])
	if !deleted {
		return val, false
	}

	// Clean up child node if empty
	if !e.node.hasValue && len(e.node.edges) == 0 {
		cur.edges = append(cur.edges[:idx], cur.edges[idx+1:]...)
	} else if !e.node.hasValue && len(e.node.edges) == 1 {
		// Merge child edge into current edge (path compression reclamation)
		grandchild := e.node.edges[0]
		e.label += grandchild.label
		e.node = grandchild.node
	}

	return val, true
}

// WalkPrefix iterates over all keys matching prefix in sorted order.
func (t *Tree[T]) WalkPrefix(prefix string, fn func(key string, val T) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cur := t.root
	search := prefix

	// Navigate to prefix root
	for len(search) > 0 {
		idx := -1
		for i := range cur.edges {
			if cur.edges[i].label[0] == search[0] {
				idx = i
				break
			}
		}
		if idx == -1 {
			return // Prefix not in tree
		}

		e := cur.edges[idx]
		if strings.HasPrefix(e.label, search) {
			// Edge label encompasses remaining prefix: walk sub-tree from here
			accum := prefix[:len(prefix)-len(search)] + e.label
			t.walkSubtree(e.node, accum, fn)
			return
		}
		if strings.HasPrefix(search, e.label) {
			search = search[len(e.label):]
			cur = e.node
			continue
		}
		return
	}

	t.walkSubtree(cur, prefix, fn)
}

func (t *Tree[T]) walkSubtree(n *node[T], currentKey string, fn func(string, T) bool) bool {
	if n.hasValue {
		if !fn(currentKey, n.val) {
			return false
		}
	}

	// Sort edges alphabetically for deterministic traversal
	edges := make([]edge[T], len(n.edges))
	copy(edges, n.edges)
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].label < edges[j].label
	})

	for _, e := range edges {
		if !t.walkSubtree(e.node, currentKey+e.label, fn) {
			return false
		}
	}
	return true
}

func commonPrefixLen(a, b string) int {
	maxLen := len(a)
	if len(b) < maxLen {
		maxLen = len(b)
	}
	for i := 0; i < maxLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return maxLen
}
