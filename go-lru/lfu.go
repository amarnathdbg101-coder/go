package golru

import (
	"sync"
	"time"
)

type lfuNode[K comparable, V any] struct {
	key       K
	val       V
	freq      int
	expiresAt int64
	prev      *lfuNode[K, V]
	next      *lfuNode[K, V]
}

type lfuDList[K comparable, V any] struct {
	head *lfuNode[K, V]
	tail *lfuNode[K, V]
	len  int
}

func (l *lfuDList[K, V]) pushFront(n *lfuNode[K, V]) {
	n.prev = nil
	n.next = l.head
	if l.head != nil {
		l.head.prev = n
	} else {
		l.tail = n
	}
	l.head = n
	l.len++
}

func (l *lfuDList[K, V]) remove(n *lfuNode[K, V]) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}

	n.prev = nil
	n.next = nil
	l.len--
}

func (l *lfuDList[K, V]) popTail() *lfuNode[K, V] {
	if l.tail == nil {
		return nil
	}
	oldTail := l.tail
	l.remove(oldTail)
	return oldTail
}

// LFUCache is an O(1) Least Frequently Used generic cache with per-item TTL expiration.
type LFUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*lfuNode[K, V]
	freqMap  map[int]*lfuDList[K, V]
	minFreq  int
	onEvict  OnEvictFunc[K, V]
	mu       sync.RWMutex
}

// NewLFU constructs an O(1) generic LFU cache.
func NewLFU[K comparable, V any](capacity int, onEvict OnEvictFunc[K, V]) (*LFUCache[K, V], error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &LFUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*lfuNode[K, V], capacity),
		freqMap:  make(map[int]*lfuDList[K, V]),
		minFreq:  0,
		onEvict:  onEvict,
	}, nil
}

// Get retrieves an item by key and increments its frequency counter.
func (c *LFUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero V
	node, exists := c.items[key]
	if !exists {
		return zero, false
	}

	if node.expiresAt > 0 && time.Now().UnixNano() >= node.expiresAt {
		c.removeNode(node)
		if c.onEvict != nil {
			c.onEvict(node.key, node.val)
		}
		return zero, false
	}

	c.incrementFreq(node)
	return node.val, true
}

// Set inserts or updates an item with default infinite expiration.
func (c *LFUCache[K, V]) Set(key K, val V) {
	c.SetWithTTL(key, val, 0)
}

// SetWithTTL inserts or updates an item with a specific TTL.
func (c *LFUCache[K, V]) SetWithTTL(key K, val V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt int64
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).UnixNano()
	}

	if node, exists := c.items[key]; exists {
		node.val = val
		node.expiresAt = expiresAt
		c.incrementFreq(node)
		return
	}

	// Evict LFU item if capacity reached
	if len(c.items) >= c.capacity {
		c.evictOne()
	}

	newNode := &lfuNode[K, V]{
		key:       key,
		val:       val,
		freq:      1,
		expiresAt: expiresAt,
	}

	c.items[key] = newNode
	c.addToFreq(1, newNode)
	c.minFreq = 1
}

// Delete removes a key explicitly from the cache.
func (c *LFUCache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.items[key]
	if !exists {
		return false
	}

	c.removeNode(node)
	return true
}

// Len returns current number of active items in cache.
func (c *LFUCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *LFUCache[K, V]) incrementFreq(node *lfuNode[K, V]) {
	oldFreq := node.freq
	list := c.freqMap[oldFreq]
	list.remove(node)

	if list.len == 0 {
		delete(c.freqMap, oldFreq)
		if c.minFreq == oldFreq {
			c.minFreq++
		}
	}

	node.freq++
	c.addToFreq(node.freq, node)
}

func (c *LFUCache[K, V]) addToFreq(freq int, node *lfuNode[K, V]) {
	list, exists := c.freqMap[freq]
	if !exists {
		list = &lfuDList[K, V]{}
		c.freqMap[freq] = list
	}
	list.pushFront(node)
}

func (c *LFUCache[K, V]) removeNode(node *lfuNode[K, V]) {
	delete(c.items, node.key)
	list := c.freqMap[node.freq]
	if list != nil {
		list.remove(node)
		if list.len == 0 {
			delete(c.freqMap, node.freq)
		}
	}
}

func (c *LFUCache[K, V]) evictOne() {
	list := c.freqMap[c.minFreq]
	if list == nil || list.len == 0 {
		return
	}

	evicted := list.popTail()
	if evicted != nil {
		delete(c.items, evicted.key)
		if c.onEvict != nil {
			c.onEvict(evicted.key, evicted.val)
		}
	}
}
