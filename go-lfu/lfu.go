package lfu

import (
	"sync"
	"time"
)

type node[K comparable, V any] struct {
	key       K
	value     V
	freq      int
	expiresAt int64 // UnixNano, 0 for no expiration
	prev      *node[K, V]
	next      *node[K, V]
}

type nodeDList[K comparable, V any] struct {
	head *node[K, V]
	tail *node[K, V]
	len  int
}

func (l *nodeDList[K, V]) pushFront(n *node[K, V]) {
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

func (l *nodeDList[K, V]) remove(n *node[K, V]) {
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

func (l *nodeDList[K, V]) popTail() *node[K, V] {
	if l.tail == nil {
		return nil
	}
	oldTail := l.tail
	l.remove(oldTail)
	return oldTail
}

// Cache is a thread-safe O(1) Least Frequently Used (LFU) cache with per-item TTL.
type Cache[K comparable, V any] struct {
	mu       sync.RWMutex
	capacity int
	minFreq  int
	items    map[K]*node[K, V]
	freqMap  map[int]*nodeDList[K, V]
}

// New creates a new LFU Cache with the given capacity.
func New[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity <= 0 {
		panic("lfu: capacity must be greater than 0")
	}
	return &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]*node[K, V]),
		freqMap:  make(map[int]*nodeDList[K, V]),
	}
}

// Get retrieves a key's value and increments its access frequency in O(1) time.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, found := c.items[key]
	if !found {
		var zero V
		return zero, false
	}

	// Check expiration
	if n.expiresAt > 0 && time.Now().UnixNano() > n.expiresAt {
		c.removeNode(n)
		var zero V
		return zero, false
	}

	c.incrementFreq(n)
	return n.value, true
}

// Set stores a key-value pair in the cache with no expiration.
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, 0)
}

// SetWithTTL stores a key-value pair with a specific time-to-live duration.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt int64
	if ttl > 0 {
		expiresAt = time.Now().UnixNano() + ttl.Nanoseconds()
	}

	// Update existing key
	if n, found := c.items[key]; found {
		n.value = value
		n.expiresAt = expiresAt
		c.incrementFreq(n)
		return
	}

	// Evict if at capacity
	if len(c.items) >= c.capacity {
		c.evictLFU()
	}

	// Insert new node with freq = 1
	newNode := &node[K, V]{
		key:       key,
		value:     value,
		freq:      1,
		expiresAt: expiresAt,
	}

	c.items[key] = newNode
	c.getOrCreateList(1).pushFront(newNode)
	c.minFreq = 1
}

// Delete removes a key from the cache.
func (c *Cache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, found := c.items[key]
	if !found {
		return false
	}

	c.removeNode(n)
	return true
}

// Len returns the number of active items in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Capacity returns the maximum capacity of the cache.
func (c *Cache[K, V]) Capacity() int {
	return c.capacity
}

// Clear removes all elements from the cache.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*node[K, V])
	c.freqMap = make(map[int]*nodeDList[K, V])
	c.minFreq = 0
}

// PurgeExpired synchronously evicts all expired items from the cache.
func (c *Cache[K, V]) PurgeExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	evicted := 0

	for _, n := range c.items {
		if n.expiresAt > 0 && now > n.expiresAt {
			c.removeNode(n)
			evicted++
		}
	}

	return evicted
}

func (c *Cache[K, V]) incrementFreq(n *node[K, V]) {
	oldFreq := n.freq
	oldList := c.freqMap[oldFreq]
	oldList.remove(n)

	if oldList.len == 0 {
		delete(c.freqMap, oldFreq)
		if c.minFreq == oldFreq {
			c.minFreq = oldFreq + 1
		}
	}

	n.freq++
	c.getOrCreateList(n.freq).pushFront(n)
}

func (c *Cache[K, V]) evictLFU() {
	minList := c.freqMap[c.minFreq]
	if minList == nil || minList.len == 0 {
		return
	}

	evicted := minList.popTail()
	if evicted != nil {
		delete(c.items, evicted.key)
		if minList.len == 0 {
			delete(c.freqMap, c.minFreq)
		}
	}
}

func (c *Cache[K, V]) removeNode(n *node[K, V]) {
	delete(c.items, n.key)
	list := c.freqMap[n.freq]
	if list != nil {
		list.remove(n)
		if list.len == 0 {
			delete(c.freqMap, n.freq)
		}
	}
}

func (c *Cache[K, V]) getOrCreateList(freq int) *nodeDList[K, V] {
	l, exists := c.freqMap[freq]
	if !exists {
		l = &nodeDList[K, V]{}
		c.freqMap[freq] = l
	}
	return l
}
