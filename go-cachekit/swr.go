package cachekit

import (
	"context"
	"encoding/json"
	"time"
)

// swrEnvelope wraps cached data with creation timestamp and soft TTL for Stale-While-Revalidate pattern.
type swrEnvelope[T any] struct {
	Value      T         `json:"value"`
	FreshUntil time.Time `json:"fresh_until"`
}

// GetOrSetSWR implements the Stale-While-Revalidate pattern:
// 1. If key is fresh -> Return immediately (0ms).
// 2. If key is stale (soft TTL passed but hard TTL remains) -> Return stale data instantly (0ms) AND launch background goroutine to refresh cache from fetcher!
// 3. If key is missing -> Fetch synchronously with Singleflight protection.
func GetOrSetSWR[T any](
	ctx context.Context,
	e *Engine,
	key string,
	freshTTL time.Duration,
	staleTTL time.Duration,
	fetcher func(ctx context.Context) (T, error),
) (T, error) {
	var zero T

	// 1. Try reading SWR envelope from cache
	rawBytes, err := e.store.Get(ctx, key)
	if err == nil {
		var env swrEnvelope[T]
		if errUnmarshal := json.Unmarshal(rawBytes, &env); errUnmarshal == nil {
			now := time.Now()
			// Fresh data: return instantly
			if now.Before(env.FreshUntil) {
				return env.Value, nil
			}

			// Stale data: serve immediately to caller (0ms), refresh asynchronously in background
			go func() {
				bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				freshVal, fetchErr := fetcher(bgCtx)
				if fetchErr == nil {
					_ = saveSWREnvelope(bgCtx, e, key, freshVal, freshTTL, staleTTL)
				}
			}()

			return env.Value, nil
		}
	}

	// 2. Cache miss: Synchronously fetch via Singleflight & save SWR envelope
	result, err, _ := e.group.Do(key, func() (any, error) {
		freshVal, fetchErr := fetcher(ctx)
		if fetchErr != nil {
			return zero, fetchErr
		}
		_ = saveSWREnvelope(ctx, e, key, freshVal, freshTTL, staleTTL)
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

func saveSWREnvelope[T any](ctx context.Context, e *Engine, key string, val T, freshTTL, staleTTL time.Duration) error {
	env := swrEnvelope[T]{
		Value:      val,
		FreshUntil: time.Now().Add(freshTTL),
	}
	bytes, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return e.store.Set(ctx, key, bytes, freshTTL+staleTTL)
}
