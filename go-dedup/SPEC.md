# Technical Specification: go-dedup

## 1. Domain & Purpose
`go-dedup` is a thread-safe, generic in-memory deduplication engine for Go. Designed specifically for webhook ingest handlers, message queue subscribers (Kafka, RabbitMQ, SQS), and high-throughput API endpoints, it absorbs duplicate events in sub-microsecond time with zero database lock contention.

## 2. Architecture & Design Principles
- **Generic Key Type**: Supports any comparable key type `K comparable` (e.g. `string` UUID/EventID, `int64` OrderID, composite struct keys).
- **Atomic Check & Mark**: `CheckAndRecord()` atomically evaluates existence, tests against expiration timestamps, and updates TTL in a single atomic transaction.
- **Sliding Retention Window**: Each entry maintains a 64-bit Unix timestamp (`int64`) in nanoseconds for low memory overhead without pointer indirection.
- **Background Janitor**: Configurable background sweeper interval that periodically purges expired keys to prevent unbounded memory growth.
- **Graceful Lifecycle**: `Close()` cleanly terminates background workers and frees memory buffers.
- **Zero Heavy Dependencies**: Pure Go standard library (`sync`, `time`).

## 3. API Surface
- `type Deduplicator[K comparable] struct`
- `func New[K comparable](defaultWindow time.Duration, cleanupInterval time.Duration) *Deduplicator[K]`
- `func (d *Deduplicator[K]) CheckAndRecord(key K) bool`
- `func (d *Deduplicator[K]) CheckAndRecordWithTTL(key K, ttl time.Duration) bool`
- `func (d *Deduplicator[K]) IsDuplicate(key K) bool`
- `func (d *Deduplicator[K]) Record(key K)`
- `func (d *Deduplicator[K]) RecordWithTTL(key K, ttl time.Duration)`
- `func (d *Deduplicator[K]) Remove(key K)`
- `func (d *Deduplicator[K]) Count() int`
- `func (d *Deduplicator[K]) Clear()`
- `func (d *Deduplicator[K]) PurgeExpired() int`
- `func (d *Deduplicator[K]) Close()`

## 4. Error Handling
- `ErrClosed`: Returned when attempting operations on a closed deduplicator.

## 5. Verification & Testing
- Unit tests verify duplicate detection, TTL expiration, per-key overrides, background janitor cleanup, manual purge, and high-concurrency race freedom across 50 concurrent goroutines.
