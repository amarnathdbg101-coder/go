package slicekit

import (
	"crypto/rand"
	"math/big"
)

// Chunk splits a slice into multiple slices of the specified size.
// If size <= 0, returns nil.
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 || len(slice) == 0 {
		return nil
	}

	total := len(slice)
	numChunks := (total + size - 1) / size
	chunks := make([][]T, 0, numChunks)

	for i := 0; i < total; i += size {
		end := i + size
		if end > total {
			end = total
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// Flatten flattens a slice of slices into a single slice with a single pre-allocated buffer.
func Flatten[T any](slices [][]T) []T {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}

	return result
}

// Unique returns a new slice containing only unique elements from the input slice,
// preserving their original first-occurrence order.
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// UniqueBy returns a new slice deduplicated by a custom key selector function.
func UniqueBy[T any, K comparable](slice []T, keyFn func(T) K) []T {
	seen := make(map[K]struct{}, len(slice))
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		key := keyFn(item)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// GroupBy groups elements of a slice into a map keyed by the result of keyFn.
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, item := range slice {
		key := keyFn(item)
		groups[key] = append(groups[key], item)
	}
	return groups
}

// Partition splits a slice into two slices: matched elements and unmatched elements.
func Partition[T any](slice []T, predicate func(T) bool) ([]T, []T) {
	var matched []T
	var unmatched []T

	for _, item := range slice {
		if predicate(item) {
			matched = append(matched, item)
		} else {
			unmatched = append(unmatched, item)
		}
	}

	return matched, unmatched
}

// Reverse returns a new slice with elements in reversed order without mutating the input.
func Reverse[T any](slice []T) []T {
	n := len(slice)
	res := make([]T, n)
	for i := 0; i < n; i++ {
		res[i] = slice[n-1-i]
	}
	return res
}

// ReverseInPlace reverses the given slice in-place.
func ReverseInPlace[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Compact removes zero-value elements from the slice.
func Compact[T comparable](slice []T) []T {
	var zero T
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if item != zero {
			result = append(result, item)
		}
	}
	return result
}

// Filter returns a new slice containing only elements that satisfy the predicate.
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map transforms elements of a slice of type T into a new slice of type R.
func Map[T any, R any](slice []T, transform func(T) R) []R {
	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = transform(item)
	}
	return result
}

// Reduce reduces a slice to a single value using a reducer function.
func Reduce[T any, R any](slice []T, initial R, reducer func(R, T) R) R {
	acc := initial
	for _, item := range slice {
		acc = reducer(acc, item)
	}
	return acc
}

// Find returns the first element satisfying the predicate, or false if not found.
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex returns the index of the first element satisfying the predicate, or -1 if not found.
func FindIndex[T any](slice []T, predicate func(T) bool) int {
	for i, item := range slice {
		if predicate(item) {
			return i
		}
	}
	return -1
}

// ContainsBy returns true if at least one element satisfies the predicate.
func ContainsBy[T any](slice []T, predicate func(T) bool) bool {
	return FindIndex(slice, predicate) != -1
}

// Shuffle randomizes the order of elements in-place using crypto/rand Fisher-Yates shuffle.
func Shuffle[T any](slice []T) {
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		bigJ, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		j := int(bigJ.Int64())
		slice[i], slice[j] = slice[j], slice[i]
	}
}
