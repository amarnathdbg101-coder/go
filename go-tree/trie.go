package gotree

import (
	"sync"
)

type trieNode[T any] struct {
	children map[rune]*trieNode[T]
	isEnd    bool
	val      T
}

func newTrieNode[T any]() *trieNode[T] {
	return &trieNode[T]{
		children: make(map[rune]*trieNode[T]),
	}
}

// Trie provides a thread-safe generic prefix tree for rapid string lookups and autocompletion.
type Trie[T any] struct {
	root *trieNode[T]
	mu   sync.RWMutex
}

// NewTrie initializes an empty Trie.
func NewTrie[T any]() *Trie[T] {
	return &Trie[T]{
		root: newTrieNode[T](),
	}
}

// Insert adds a string key with an associated generic payload.
func (t *Trie[T]) Insert(key string, val T) {
	t.mu.Lock()
	defer t.mu.Unlock()

	curr := t.root
	for _, r := range key {
		child, exists := curr.children[r]
		if !exists {
			child = newTrieNode[T]()
			curr.children[r] = child
		}
		curr = child
	}
	curr.isEnd = true
	curr.val = val
}

// Search checks if exact key exists and returns its value.
func (t *Trie[T]) Search(key string) (T, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var zero T
	curr := t.root
	for _, r := range key {
		child, exists := curr.children[r]
		if !exists {
			return zero, false
		}
		curr = child
	}

	if curr.isEnd {
		return curr.val, true
	}
	return zero, false
}

// StartsWith retrieves all values matching given prefix.
func (t *Trie[T]) StartsWith(prefix string) []T {
	t.mu.RLock()
	defer t.mu.RUnlock()

	curr := t.root
	for _, r := range prefix {
		child, exists := curr.children[r]
		if !exists {
			return nil
		}
		curr = child
	}

	var results []T
	var collect func(node *trieNode[T])
	collect = func(node *trieNode[T]) {
		if node.isEnd {
			results = append(results, node.val)
		}
		for _, child := range node.children {
			collect(child)
		}
	}

	collect(curr)
	return results
}

// Delete removes a key from the trie.
func (t *Trie[T]) Delete(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	runes := []rune(key)
	return t.deleteHelper(t.root, runes, 0)
}

func (t *Trie[T]) deleteHelper(curr *trieNode[T], runes []rune, depth int) bool {
	if curr == nil {
		return false
	}

	if depth == len(runes) {
		if !curr.isEnd {
			return false
		}
		curr.isEnd = false
		return true
	}

	r := runes[depth]
	child, exists := curr.children[r]
	if !exists {
		return false
	}

	deleted := t.deleteHelper(child, runes, depth+1)
	if deleted && !child.isEnd && len(child.children) == 0 {
		delete(curr.children, r)
	}

	return deleted
}
