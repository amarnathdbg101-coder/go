package dataloader

import (
	"context"
	"sync"
	"time"
)

// BatchFunc is the user-supplied function that resolves a batch of keys in a single round-trip.
// It must return a slice of Result[V] having the exact same length and order as keys.
type BatchFunc[K comparable, V any] func(ctx context.Context, keys []K) ([]Result[V], error)

type pendingBatch[K comparable, V any] struct {
	keys       []K
	keyIndex   map[K]int
	ch         chan struct{}
	results    []Result[V]
	batchErr   error
	timer      *time.Timer
	dispatched bool
}

// Loader is a generic batching and caching data loader for Go.
type Loader[K comparable, V any] struct {
	batchFn BatchFunc[K, V]
	cfg     Config

	mu           sync.Mutex
	cache        map[K]Result[V]
	currentBatch *pendingBatch[K, V]
	closed       bool
}

// New creates a new DataLoader with the given batch function and options.
func New[K comparable, V any](batchFn BatchFunc[K, V], opts ...Option) *Loader[K, V] {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Loader[K, V]{
		batchFn: batchFn,
		cfg:     cfg,
		cache:   make(map[K]Result[V]),
	}
}

// Load requests a single value by key. If multiple goroutines request keys concurrently,
// they are coalesced into a single batched call.
func (l *Loader[K, V]) Load(ctx context.Context, key K) (V, error) {
	var zero V

	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return zero, ErrLoaderClosed
	}

	// 1. Check memoization cache
	if l.cfg.Cache {
		if res, found := l.cache[key]; found {
			l.mu.Unlock()
			return res.Data, res.Err
		}
	}

	// 2. Add to or initialize current batch
	b := l.currentBatch
	if b == nil {
		b = &pendingBatch[K, V]{
			keys:     make([]K, 0, l.cfg.MaxBatch),
			keyIndex: make(map[K]int),
			ch:       make(chan struct{}),
		}
		l.currentBatch = b

		// Schedule dispatch after Wait duration
		b.timer = time.AfterFunc(l.cfg.Wait, func() {
			l.dispatch(b)
		})
	}

	// Check if key already in this batch
	idx, exists := b.keyIndex[key]
	if !exists {
		idx = len(b.keys)
		b.keys = append(b.keys, key)
		b.keyIndex[key] = idx

		// If max batch size reached, trigger immediately
		if len(b.keys) >= l.cfg.MaxBatch {
			b.timer.Stop()
			go l.dispatch(b)
		}
	}

	waitChan := b.ch
	l.mu.Unlock()

	// 3. Wait for batch completion or context cancellation
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-waitChan:
		// Batch finished
		if b.batchErr != nil {
			return zero, b.batchErr
		}
		if idx >= len(b.results) {
			return zero, ErrResultLengthMismatch
		}
		res := b.results[idx]
		return res.Data, res.Err
	}
}

// LoadMany loads multiple keys concurrently, returning results in matching order.
func (l *Loader[K, V]) LoadMany(ctx context.Context, keys []K) ([]V, []error) {
	results := make([]V, len(keys))
	errors := make([]error, len(keys))

	var wg sync.WaitGroup
	wg.Add(len(keys))

	for i, k := range keys {
		go func(idx int, key K) {
			defer wg.Done()
			val, err := l.Load(ctx, key)
			results[idx] = val
			errors[idx] = err
		}(i, k)
	}

	wg.Wait()
	return results, errors
}

// Prime pre-warms the cache with a key-value pair.
func (l *Loader[K, V]) Prime(key K, value V) {
	if !l.cfg.Cache {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache[key] = Result[V]{Data: value, Err: nil}
}

// Clear removes a single key from the cache.
func (l *Loader[K, V]) Clear(key K) {
	if !l.cfg.Cache {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.cache, key)
}

// ClearAll evicts all entries from the cache.
func (l *Loader[K, V]) ClearAll() {
	if !l.cfg.Cache {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache = make(map[K]Result[V])
}

// Close closes the loader and clears internal structures.
func (l *Loader[K, V]) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
	if l.currentBatch != nil && l.currentBatch.timer != nil {
		l.currentBatch.timer.Stop()
	}
	l.cache = nil
	return nil
}

// dispatch runs the batch function and distributes results.
func (l *Loader[K, V]) dispatch(b *pendingBatch[K, V]) {
	l.mu.Lock()
	if b.dispatched {
		l.mu.Unlock()
		return
	}
	b.dispatched = true
	if l.currentBatch == b {
		l.currentBatch = nil
	}
	l.mu.Unlock()

	// Execute user batch function with background context or dedicated timeout
	ctx := context.Background()
	results, err := l.batchFn(ctx, b.keys)
	b.batchErr = err
	b.results = results

	// Cache valid results if caching is enabled
	if l.cfg.Cache && err == nil && len(results) == len(b.keys) {
		l.mu.Lock()
		for i, key := range b.keys {
			l.cache[key] = results[i]
		}
		l.mu.Unlock()
	}

	// Notify all waiting callers
	close(b.ch)
}
