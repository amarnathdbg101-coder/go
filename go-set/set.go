package set

import (
	"encoding/json"
	"sync"
)

// Set is a generic collection of unique elements of type T.
type Set[T comparable] struct {
	mu    *sync.RWMutex
	items map[T]struct{}
}

// New creates a new non-thread-safe Set populated with the given items.
func New[T comparable](items ...T) *Set[T] {
	s := &Set[T]{
		items: make(map[T]struct{}, len(items)),
	}
	s.Add(items...)
	return s
}

// NewThreadSafe creates a new thread-safe Set protected by an internal RWMutex.
func NewThreadSafe[T comparable](items ...T) *Set[T] {
	s := &Set[T]{
		mu:    &sync.RWMutex{},
		items: make(map[T]struct{}, len(items)),
	}
	s.Add(items...)
	return s
}

func (s *Set[T]) lock() {
	if s.mu != nil {
		s.mu.Lock()
	}
}

func (s *Set[T]) unlock() {
	if s.mu != nil {
		s.mu.Unlock()
	}
}

func (s *Set[T]) rlock() {
	if s.mu != nil {
		s.mu.RLock()
	}
}

func (s *Set[T]) runlock() {
	if s.mu != nil {
		s.mu.RUnlock()
	}
}

// Add adds one or more elements to the set.
func (s *Set[T]) Add(items ...T) {
	s.lock()
	defer s.unlock()

	if s.items == nil {
		s.items = make(map[T]struct{})
	}
	for _, item := range items {
		s.items[item] = struct{}{}
	}
}

// Remove deletes one or more elements from the set.
func (s *Set[T]) Remove(items ...T) {
	s.lock()
	defer s.unlock()

	for _, item := range items {
		delete(s.items, item)
	}
}

// Contains returns true if the set contains the given item.
func (s *Set[T]) Contains(item T) bool {
	s.rlock()
	defer s.runlock()

	_, found := s.items[item]
	return found
}

// ContainsAll returns true if all given items are in the set.
func (s *Set[T]) ContainsAll(items ...T) bool {
	s.rlock()
	defer s.runlock()

	for _, item := range items {
		if _, found := s.items[item]; !found {
			return false
		}
	}
	return true
}

// ContainsAny returns true if at least one of the given items is in the set.
func (s *Set[T]) ContainsAny(items ...T) bool {
	s.rlock()
	defer s.runlock()

	for _, item := range items {
		if _, found := s.items[item]; found {
			return true
		}
	}
	return false
}

// Size returns the number of distinct elements in the set.
func (s *Set[T]) Size() int {
	s.rlock()
	defer s.runlock()
	return len(s.items)
}

// IsEmpty returns true if the set has no elements.
func (s *Set[T]) IsEmpty() bool {
	return s.Size() == 0
}

// Clear removes all elements from the set.
func (s *Set[T]) Clear() {
	s.lock()
	defer s.unlock()
	s.items = make(map[T]struct{})
}

// ToSlice returns all elements in the set as a slice.
func (s *Set[T]) ToSlice() []T {
	s.rlock()
	defer s.runlock()

	slice := make([]T, 0, len(s.items))
	for item := range s.items {
		slice = append(slice, item)
	}
	return slice
}

// Clone creates an independent copy of the set, preserving thread-safety settings.
func (s *Set[T]) Clone() *Set[T] {
	s.rlock()
	defer s.runlock()

	var newSet *Set[T]
	if s.mu != nil {
		newSet = NewThreadSafe[T]()
	} else {
		newSet = New[T]()
	}

	for item := range s.items {
		newSet.items[item] = struct{}{}
	}
	return newSet
}

// Equal returns true if both sets have the exact same elements.
func (s *Set[T]) Equal(other *Set[T]) bool {
	if other == nil {
		return false
	}
	if s == other {
		return true
	}

	s.rlock()
	other.rlock()
	defer s.runlock()
	defer other.runlock()

	if len(s.items) != len(other.items) {
		return false
	}

	for item := range s.items {
		if _, found := other.items[item]; !found {
			return false
		}
	}
	return true
}

// Union returns a new set containing all elements present in either set.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	res := s.Clone()
	if other == nil {
		return res
	}

	other.rlock()
	defer other.runlock()

	for item := range other.items {
		res.items[item] = struct{}{}
	}
	return res
}

// Intersect returns a new set containing only elements present in both sets.
func (s *Set[T]) Intersect(other *Set[T]) *Set[T] {
	var res *Set[T]
	if s.mu != nil {
		res = NewThreadSafe[T]()
	} else {
		res = New[T]()
	}

	if other == nil {
		return res
	}

	s.rlock()
	other.rlock()
	defer s.runlock()
	defer other.runlock()

	for item := range s.items {
		if _, found := other.items[item]; found {
			res.items[item] = struct{}{}
		}
	}
	return res
}

// Difference returns a new set containing elements in s that are not in other.
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	res := s.Clone()
	if other == nil {
		return res
	}

	other.rlock()
	defer other.runlock()

	for item := range other.items {
		delete(res.items, item)
	}
	return res
}

// SymmetricDifference returns a new set containing elements in either s or other, but not both.
func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T] {
	u := s.Union(other)
	i := s.Intersect(other)
	return u.Difference(i)
}

// IsSubset returns true if all elements of s are also in other.
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	if other == nil {
		return false
	}

	s.rlock()
	other.rlock()
	defer s.runlock()
	defer other.runlock()

	if len(s.items) > len(other.items) {
		return false
	}

	for item := range s.items {
		if _, found := other.items[item]; !found {
			return false
		}
	}
	return true
}

// IsSuperset returns true if s contains all elements of other.
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	if other == nil {
		return true
	}
	return other.IsSubset(s)
}

// Filter returns a new set containing elements that satisfy the predicate.
func (s *Set[T]) Filter(predicate func(T) bool) *Set[T] {
	s.rlock()
	defer s.runlock()

	var res *Set[T]
	if s.mu != nil {
		res = NewThreadSafe[T]()
	} else {
		res = New[T]()
	}

	for item := range s.items {
		if predicate(item) {
			res.items[item] = struct{}{}
		}
	}
	return res
}

// Each calls fn on every element in the set.
func (s *Set[T]) Each(fn func(T)) {
	s.rlock()
	defer s.runlock()

	for item := range s.items {
		fn(item)
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (s *Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ToSlice())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	var slice []T
	if err := json.Unmarshal(data, &slice); err != nil {
		return err
	}

	s.lock()
	defer s.unlock()

	s.items = make(map[T]struct{}, len(slice))
	for _, item := range slice {
		s.items[item] = struct{}{}
	}
	return nil
}
