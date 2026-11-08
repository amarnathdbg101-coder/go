package goconcurrentmap

import (
	"fmt"
	"hash/fnv"
	"sync"
)

type shard[K comparable, V any] struct {
	items map[K]V
	mu    sync.RWMutex
}

// ConcurrentMap implements a high-throughput, 32-way striped concurrent hash map.
type ConcurrentMap[K comparable, V any] struct {
	shards []*shard[K, V]
	mask   uint32
}

// New creates a ConcurrentMap with default 32 shards.
func New[K comparable, V any]() *ConcurrentMap[K, V] {
	return NewWithShards[K, V](32)
}

// NewWithShards creates a ConcurrentMap with specified power-of-two shard count.
func NewWithShards[K comparable, V any](shardCount int) *ConcurrentMap[K, V] {
	if shardCount <= 0 {
		shardCount = 32
	}
	// Round up to power of two
	power := 1
	for power < shardCount {
		power <<= 1
	}

	cm := &ConcurrentMap[K, V]{
		shards: make([]*shard[K, V], power),
		mask:   uint32(power - 1),
	}

	for i := 0; i < power; i++ {
		cm.shards[i] = &shard[K, V]{
			items: make(map[K]V),
		}
	}

	return cm
}

func (cm *ConcurrentMap[K, V]) getShard(key K) *shard[K, V] {
	var keyBytes []byte
	switch k := any(key).(type) {
	case string:
		keyBytes = []byte(k)
	default:
		keyBytes = []byte(fmt.Sprintf("%v", key))
	}

	h := fnv.New32a()
	h.Write(keyBytes)
	idx := h.Sum32() & cm.mask
	return cm.shards[idx]
}

// Set stores a key-value pair.
func (cm *ConcurrentMap[K, V]) Set(key K, val V) {
	sh := cm.getShard(key)
	sh.mu.Lock()
	sh.items[key] = val
	sh.mu.Unlock()
}

// Get retrieves a value by key.
func (cm *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	sh := cm.getShard(key)
	sh.mu.RLock()
	val, ok := sh.items[key]
	sh.mu.RUnlock()
	return val, ok
}

// GetOrSet returns existing value or stores and returns newVal.
func (cm *ConcurrentMap[K, V]) GetOrSet(key K, newVal V) (V, bool) {
	sh := cm.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	if val, ok := sh.items[key]; ok {
		return val, true
	}

	sh.items[key] = newVal
	return newVal, false
}

// Has checks if key exists.
func (cm *ConcurrentMap[K, V]) Has(key K) bool {
	sh := cm.getShard(key)
	sh.mu.RLock()
	_, ok := sh.items[key]
	sh.mu.RUnlock()
	return ok
}

// Remove deletes a key from map.
func (cm *ConcurrentMap[K, V]) Remove(key K) bool {
	sh := cm.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	if _, ok := sh.items[key]; ok {
		delete(sh.items, key)
		return true
	}
	return false
}

// Pop deletes and returns the key's value.
func (cm *ConcurrentMap[K, V]) Pop(key K) (V, bool) {
	sh := cm.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	var zero V
	if val, ok := sh.items[key]; ok {
		delete(sh.items, key)
		return val, true
	}
	return zero, false
}

// Upsert updates an entry atomically using callback function.
func (cm *ConcurrentMap[K, V]) Upsert(key K, val V, cb func(exists bool, oldVal V, newVal V) V) V {
	sh := cm.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	oldVal, exists := sh.items[key]
	finalVal := cb(exists, oldVal, val)
	sh.items[key] = finalVal
	return finalVal
}

// Count returns total items across all shards.
func (cm *ConcurrentMap[K, V]) Count() int {
	total := 0
	for _, sh := range cm.shards {
		sh.mu.RLock()
		total += len(sh.items)
		sh.mu.RUnlock()
	}
	return total
}

// IsEmpty returns true if map has 0 elements.
func (cm *ConcurrentMap[K, V]) IsEmpty() bool {
	return cm.Count() == 0
}

// Clear removes all keys from all shards.
func (cm *ConcurrentMap[K, V]) Clear() {
	for _, sh := range cm.shards {
		sh.mu.Lock()
		sh.items = make(map[K]V)
		sh.mu.Unlock()
	}
}

// Keys returns a slice of all keys.
func (cm *ConcurrentMap[K, V]) Keys() []K {
	var keys []K
	for _, sh := range cm.shards {
		sh.mu.RLock()
		for k := range sh.items {
			keys = append(keys, k)
		}
		sh.mu.RUnlock()
	}
	return keys
}

// Values returns a slice of all values.
func (cm *ConcurrentMap[K, V]) Values() []V {
	var values []V
	for _, sh := range cm.shards {
		sh.mu.RLock()
		for _, v := range sh.items {
			values = append(values, v)
		}
		sh.mu.RUnlock()
	}
	return values
}

// Iter iterates through all items. Returning false halts iteration early.
func (cm *ConcurrentMap[K, V]) Iter(fn func(key K, val V) bool) {
	for _, sh := range cm.shards {
		sh.mu.RLock()
		for k, v := range sh.items {
			if !fn(k, v) {
				sh.mu.RUnlock()
				return
			}
		}
		sh.mu.RUnlock()
	}
}
