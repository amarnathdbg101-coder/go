// Package dataloader provides high-performance, generic batching and caching for Go.
// It coalesces multiple concurrent lookups into single batched calls to completely eliminate N+1 query bottlenecks.
package dataloader

import "context"

// Version of the go-dataloader package.
const Version = "1.0.0"

// BatchFromMap is a helper function that converts a map lookup function into a BatchFunc.
func BatchFromMap[K comparable, V any](fn func(ctx context.Context, keys []K) (map[K]V, error)) BatchFunc[K, V] {
	return func(ctx context.Context, keys []K) ([]Result[V], error) {
		m, err := fn(ctx, keys)
		if err != nil {
			return nil, err
		}
		results := make([]Result[V], len(keys))
		for i, k := range keys {
			if val, ok := m[k]; ok {
				results[i] = Result[V]{Data: val}
			} else {
				results[i] = Result[V]{Err: ErrKeyNotFound}
			}
		}
		return results, nil
	}
}
