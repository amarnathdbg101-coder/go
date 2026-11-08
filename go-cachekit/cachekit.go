package cachekit

import (
	"context"
	"encoding/json"
	"time"
)

// Engine is the central caching orchestrator.
// Isme storage backend, singleflight coalescing, aur serialization logic encapsulate hoti hai.
type Engine struct {
	store Store
	group *Group
}

// New creates and initializes a ready-to-use cachekit Engine.
func New(store Store) (*Engine, error) {
	if store == nil {
		return nil, ErrStoreNil
	}
	return &Engine{
		store: store,
		group: NewGroup(),
	}, nil
}

// Store returns the underlying storage implementation.
func (e *Engine) Store() Store {
	return e.store
}

// GetBytes retrieves raw cached bytes.
func (e *Engine) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return e.store.Get(ctx, key)
}

// SetBytes saves raw bytes with specified TTL.
func (e *Engine) SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return e.store.Set(ctx, key, value, ttl)
}

// Delete evicts a specific key from cache.
func (e *Engine) Delete(ctx context.Context, key string) error {
	return e.store.Delete(ctx, key)
}

// DeletePrefix evicts all keys matching a prefix (e.g. "products:").
func (e *Engine) DeletePrefix(ctx context.Context, prefix string) error {
	return e.store.DeletePrefix(ctx, prefix)
}

// Clear flushes all entries.
func (e *Engine) Clear(ctx context.Context) error {
	return e.store.Clear(ctx)
}

// Close releases engine and store resources.
func (e *Engine) Close() error {
	return e.store.Close()
}

// -------------------------------------------------------------------------
// Generics-Powered High Level Functions
// -------------------------------------------------------------------------

// Get retrieves and unmarshals JSON cached data into strongly-typed value T.
func Get[T any](ctx context.Context, e *Engine, key string) (T, error) {
	var zero T
	bytes, err := e.store.Get(ctx, key)
	if err != nil {
		return zero, err
	}

	var val T
	if err := json.Unmarshal(bytes, &val); err != nil {
		return zero, err
	}
	return val, nil
}

// Set serializes value T into JSON bytes and caches it with TTL.
func Set[T any](ctx context.Context, e *Engine, key string, value T, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return e.store.Set(ctx, key, bytes, ttl)
}

// GetOrSet handles the complete read-through cache pattern with Dogpile / Cache Stampede protection:
// 1. Checks cache for key. If hit, unmarshals and returns immediately.
// 2. If miss, uses Singleflight group to coalesce concurrent requests into exactly ONE execution.
// 3. Runs fetcher function (e.g. database query) only once.
// 4. Writes result to cache with specified TTL.
// 5. Returns typed result to all concurrent waiting callers.
func GetOrSet[T any](
	ctx context.Context,
	e *Engine,
	key string,
	ttl time.Duration,
	fetcher func(ctx context.Context) (T, error),
) (T, error) {
	var zero T

	// 1. Fast path: check cache first
	bytes, err := e.store.Get(ctx, key)
	if err == nil {
		var cachedVal T
		if err := json.Unmarshal(bytes, &cachedVal); err == nil {
			return cachedVal, nil
		}
	}

	// 2. Slow path: Coalesce via Singleflight to prevent DB crash
	result, err, _ := e.group.Do(key, func() (any, error) {
		// Double-check cache inside singleflight lock (in case a prior call populated it)
		if b, errCheck := e.store.Get(ctx, key); errCheck == nil {
			var cachedVal T
			if errUnmarshal := json.Unmarshal(b, &cachedVal); errUnmarshal == nil {
				return cachedVal, nil
			}
		}

		// Execute actual DB fetch
		freshVal, fetchErr := fetcher(ctx)
		if fetchErr != nil {
			return zero, fetchErr
		}

		// Serialize and store into cache
		if dataBytes, marshalErr := json.Marshal(freshVal); marshalErr == nil {
			_ = e.store.Set(ctx, key, dataBytes, ttl)
		}

		return freshVal, nil
	})

	if err != nil {
		return zero, err
	}

	typed, ok := result.(T)
	if !ok {
		return zero, ErrCacheMiss
	}
	return typed, nil
}

// -------------------------------------------------------------------------
// Typed Cache Wrapper (Domain-specific ergonomics)
// -------------------------------------------------------------------------

// TypedCache provides a type-safe wrapper for a specific entity type (e.g. Product, User).
type TypedCache[T any] struct {
	engine *Engine
}

// Typed creates a type-bound cache client for struct T.
//
// Example:
//
//	productCache := cachekit.Typed[Product](engine)
//	product, err := productCache.GetOrSet(ctx, "prod:101", 10*time.Minute, fetchProductFromPostgres)
func Typed[T any](engine *Engine) *TypedCache[T] {
	return &TypedCache[T]{engine: engine}
}

// Get retrieves cached item.
func (tc *TypedCache[T]) Get(ctx context.Context, key string) (T, error) {
	return Get[T](ctx, tc.engine, key)
}

// Set stores item in cache.
func (tc *TypedCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	return Set[T](ctx, tc.engine, key, value, ttl)
}

// GetOrSet retrieves from cache or invokes fetcher with singleflight stampede protection.
func (tc *TypedCache[T]) GetOrSet(ctx context.Context, key string, ttl time.Duration, fetcher func(ctx context.Context) (T, error)) (T, error) {
	return GetOrSet[T](ctx, tc.engine, key, ttl, fetcher)
}

// Delete evicts key.
func (tc *TypedCache[T]) Delete(ctx context.Context, key string) error {
	return tc.engine.Delete(ctx, key)
}
