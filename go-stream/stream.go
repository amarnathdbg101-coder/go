package gostream

import (
	"sort"
	"sync"
)

// Stream encapsulates a sequence of elements supporting functional operations.
type Stream[T any] struct {
	items []T
}

// FromSlice constructs a Stream from an existing slice.
func FromSlice[T any](items []T) *Stream[T] {
	copied := make([]T, len(items))
	copy(copied, items)
	return &Stream[T]{items: copied}
}

// Empty constructs an empty Stream.
func Empty[T any]() *Stream[T] {
	return &Stream[T]{items: make([]T, 0)}
}

// Filter retains only elements matching the predicate.
func (s *Stream[T]) Filter(predicate func(T) bool) *Stream[T] {
	result := make([]T, 0, len(s.items))
	for _, item := range s.items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return &Stream[T]{items: result}
}

// Distinct filters out duplicates based on a key selector.
func (s *Stream[T]) Distinct(keySelector func(T) any) *Stream[T] {
	seen := make(map[any]struct{}, len(s.items))
	result := make([]T, 0, len(s.items))

	for _, item := range s.items {
		key := keySelector(item)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}
	return &Stream[T]{items: result}
}

// Sort orders the stream using a custom comparator.
func (s *Stream[T]) Sort(less func(a, b T) bool) *Stream[T] {
	result := make([]T, len(s.items))
	copy(result, s.items)

	sort.SliceStable(result, func(i, j int) bool {
		return less(result[i], result[j])
	})
	return &Stream[T]{items: result}
}

// Take returns a stream with the first n elements.
func (s *Stream[T]) Take(n int) *Stream[T] {
	if n <= 0 {
		return Empty[T]()
	}
	if n >= len(s.items) {
		return s
	}
	return &Stream[T]{items: s.items[:n]}
}

// Skip skips the first n elements.
func (s *Stream[T]) Skip(n int) *Stream[T] {
	if n <= 0 {
		return s
	}
	if n >= len(s.items) {
		return Empty[T]()
	}
	return &Stream[T]{items: s.items[n:]}
}

// ToSlice evaluates the stream and collects elements into a slice.
func (s *Stream[T]) ToSlice() []T {
	result := make([]T, len(s.items))
	copy(result, s.items)
	return result
}

// ForEach executes an action for each element.
func (s *Stream[T]) ForEach(action func(T)) {
	for _, item := range s.items {
		action(item)
	}
}

// First returns the first element or (zero, false) if empty.
func (s *Stream[T]) First() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[0], true
}

// Count returns the number of elements in the stream.
func (s *Stream[T]) Count() int {
	return len(s.items)
}

// AnyMatch returns true if at least one element satisfies the predicate.
func (s *Stream[T]) AnyMatch(predicate func(T) bool) bool {
	for _, item := range s.items {
		if predicate(item) {
			return true
		}
	}
	return false
}

// AllMatch returns true if every element satisfies the predicate.
func (s *Stream[T]) AllMatch(predicate func(T) bool) bool {
	for _, item := range s.items {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Reduce aggregates elements using an accumulator function.
func (s *Stream[T]) Reduce(initial T, accumulator func(acc, item T) T) T {
	acc := initial
	for _, item := range s.items {
		acc = accumulator(acc, item)
	}
	return acc
}

// Map transforms elements of type T to type R.
func Map[T, R any](s *Stream[T], transform func(T) R) *Stream[R] {
	result := make([]R, len(s.items))
	for i, item := range s.items {
		result[i] = transform(item)
	}
	return &Stream[R]{items: result}
}

// FlatMap flattens slices produced by transform into a single stream.
func FlatMap[T, R any](s *Stream[T], transform func(T) []R) *Stream[R] {
	var result []R
	for _, item := range s.items {
		result = append(result, transform(item)...)
	}
	return &Stream[R]{items: result}
}

// Chunk partitions the stream into slices of specified size.
func Chunk[T any](s *Stream[T], size int) *Stream[[]T] {
	if size <= 0 {
		size = 100
	}
	var chunks [][]T
	for i := 0; i < len(s.items); i += size {
		end := i + size
		if end > len(s.items) {
			end = len(s.items)
		}
		chunk := make([]T, end-i)
		copy(chunk, s.items[i:end])
		chunks = append(chunks, chunk)
	}
	return &Stream[[]T]{items: chunks}
}

type parallelJob[T any] struct {
	index int
	item  T
}

// ParallelMap executes the transform concurrently across a worker pool while preserving original order.
func ParallelMap[T, R any](s *Stream[T], workers int, transform func(T) R) *Stream[R] {
	if workers <= 0 {
		workers = 4
	}
	itemCount := len(s.items)
	if itemCount == 0 {
		return Empty[R]()
	}

	results := make([]R, itemCount)
	jobs := make(chan parallelJob[T], itemCount)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results[job.index] = transform(job.item)
			}
		}()
	}

	for i, item := range s.items {
		jobs <- parallelJob[T]{index: i, item: item}
	}
	close(jobs)

	wg.Wait()
	return &Stream[R]{items: results}
}
