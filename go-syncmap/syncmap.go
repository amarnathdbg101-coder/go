package syncmap

import (
	"reflect"
	"sync"
)

// Map is a type-safe generic concurrent map wrapping sync.Map.
type Map[K comparable, V any] struct {
	m  sync.Map
	mu sync.Mutex // Used for atomic Compute transactions
}

// New creates a new type-safe concurrent Map.
func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

// Load returns the value stored for a key, or false if not found.
func (m *Map[K, V]) Load(key K) (V, bool) {
	val, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

// Store sets the value for a key.
func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

// LoadOrStore returns the existing value if present. Otherwise, it stores and returns the given value.
// loaded is true if the value was loaded, false if stored.
func (m *Map[K, V]) LoadOrStore(key K, value V) (V, bool) {
	actual, loaded := m.m.LoadOrStore(key, value)
	return actual.(V), loaded
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
func (m *Map[K, V]) LoadAndDelete(key K) (V, bool) {
	val, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		var zero V
		return zero, false
	}
	return val.(V), true
}

// Delete deletes the value for a key.
func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Swap swaps the value for a key and returns the previous value if any.
func (m *Map[K, V]) Swap(key K, value V) (V, bool) {
	prev, loaded := m.m.Swap(key, value)
	if !loaded {
		var zero V
		return zero, false
	}
	return prev.(V), true
}

// CompareAndSwap swaps the old value for new value if current value matches old.
func (m *Map[K, V]) CompareAndSwap(key K, oldVal, newVal V) bool {
	return m.m.CompareAndSwap(key, oldVal, newVal)
}

// CompareAndDelete deletes the entry for key if its value equals oldVal.
func (m *Map[K, V]) CompareAndDelete(key K, oldVal V) bool {
	return m.m.CompareAndDelete(key, oldVal)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

// Keys returns a slice of all keys currently in the map.
func (m *Map[K, V]) Keys() []K {
	var keys []K
	m.Range(func(k K, _ V) bool {
		keys = append(keys, k)
		return true
	})
	return keys
}

// Values returns a slice of all values currently in the map.
func (m *Map[K, V]) Values() []V {
	var values []V
	m.Range(func(_ K, v V) bool {
		values = append(values, v)
		return true
	})
	return values
}

// Len returns the number of items in the map.
func (m *Map[K, V]) Len() int {
	count := 0
	m.m.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// Clear removes all keys from the map.
func (m *Map[K, V]) Clear() {
	m.m.Clear()
}

// Compute atomically computes a new value for key based on the existing value.
// If fn returns delete=true, the key is removed from the map.
func (m *Map[K, V]) Compute(key K, fn func(oldValue V, exists bool) (newValue V, delete bool)) V {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldVal, exists := m.Load(key)
	newVal, del := fn(oldVal, exists)

	if del {
		m.Delete(key)
		var zero V
		return zero
	}

	m.Store(key, newVal)
	return newVal
}

// Has returns true if the key exists in the map.
func (m *Map[K, V]) Has(key K) bool {
	_, ok := m.Load(key)
	return ok
}

// Clone returns an independent snapshot copy of the map.
func (m *Map[K, V]) Clone() *Map[K, V] {
	clone := New[K, V]()
	m.Range(func(k K, v V) bool {
		clone.Store(k, v)
		return true
	})
	return clone
}

// Equal returns true if both maps have identical keys and values.
func (m *Map[K, V]) Equal(other *Map[K, V]) bool {
	if other == nil {
		return false
	}
	if m == other {
		return true
	}
	if m.Len() != other.Len() {
		return false
	}

	equal := true
	m.Range(func(k K, v V) bool {
		otherVal, ok := other.Load(k)
		if !ok || !reflect.DeepEqual(v, otherVal) {
			equal = false
			return false
		}
		return true
	})
	return equal
}
