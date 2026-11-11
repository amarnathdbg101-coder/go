# Technical Specification: go-cachekit/lfu

## 1. Domain & Purpose
`go-cachekit/lfu` is an $O(1)$ Least Frequently Used (LFU) cache engine for Go. Unlike LRU, which tracks access recency, LFU tracks access frequency, preserving long-tail hot items (such as popular search keywords, high-traffic product catalog entries, and system tokens) from eviction during temporary scan spikes.

## 2. Architecture & Design Principles
- **Constant Time $O(1)$ Operations**: Backed by a dual-hash doubly-linked list structure: a hash map of nodes and a frequency map of doubly-linked lists.
- **Tie-Breaking Eviction**: When multiple items share the minimum frequency (`minFreq`), eviction tie-breaks using recency (evicting the oldest node in that frequency list).
- **Per-Item TTL Expiration**: Supports optional time-to-live (`SetWithTTL`) with lazy on-access expiration and batch synchronous purging (`PurgeExpired`).
- **Thread Safety**: Fully synchronized using `sync.RWMutex`.
- **Zero Heavy Dependencies**: Pure standard library (`sync`, `time`).

## 3. API Surface
- `type Cache[K comparable, V any] struct`
- `func New[K comparable, V any](capacity int) *Cache[K, V]`
- `func (c *Cache[K, V]) Get(key K) (V, bool)`
- `func (c *Cache[K, V]) Set(key K, value V)`
- `func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration)`
- `func (c *Cache[K, V]) Delete(key K) bool`
- `func (c *Cache[K, V]) Len() int`
- `func (c *Cache[K, V]) Capacity() int`
- `func (c *Cache[K, V]) Clear()`
- `func (c *Cache[K, V]) PurgeExpired() int`

## 4. Error Handling
- `ErrInvalidCapacity`: Returned when capacity is $\le 0$.

## 5. Verification & Testing
- Unit tests verify $O(1)$ frequency increments, least-frequently-used eviction prioritization, FIFO tie-breaking on identical frequencies, TTL expiration, purge mechanisms, and concurrency across 40 parallel routines.
