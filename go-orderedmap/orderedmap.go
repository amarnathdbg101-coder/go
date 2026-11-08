package orderedmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// Element represents a single key-value node in the ordered map.
type Element[K comparable, V any] struct {
	Key   K
	Value V
	prev  *Element[K, V]
	next  *Element[K, V]
}

// Next returns the next element in the map, or nil if at the tail.
func (e *Element[K, V]) Next() *Element[K, V] {
	if e == nil {
		return nil
	}
	return e.next
}

// Prev returns the previous element in the map, or nil if at the head.
func (e *Element[K, V]) Prev() *Element[K, V] {
	if e == nil {
		return nil
	}
	return e.prev
}

// Entry is a lightweight key-value pair snapshot.
type Entry[K comparable, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}

// Map is a generic ordered hash map preserving insertion order.
// O(1) Get, O(1) Set, and O(1) Delete operations.
type Map[K comparable, V any] struct {
	mu         sync.RWMutex
	concurrent bool
	entries    map[K]*Element[K, V]
	head       *Element[K, V]
	tail       *Element[K, V]
}

// New creates a new non-concurrent OrderedMap.
func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		entries: make(map[K]*Element[K, V]),
	}
}

// NewConcurrent creates a thread-safe OrderedMap guarded by sync.RWMutex.
func NewConcurrent[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		concurrent: true,
		entries:    make(map[K]*Element[K, V]),
	}
}

// Set inserts or updates a key-value pair.
// If the key already exists, its value is updated without altering its insertion position.
func (m *Map[K, V]) Set(key K, value V) {
	if m.concurrent {
		m.mu.Lock()
		defer m.mu.Unlock()
	}

	if elem, ok := m.entries[key]; ok {
		elem.Value = value
		return
	}

	elem := &Element[K, V]{
		Key:   key,
		Value: value,
		prev:  m.tail,
	}

	if m.tail != nil {
		m.tail.next = elem
	} else {
		m.head = elem
	}
	m.tail = elem
	m.entries[key] = elem
}

// Get retrieves the value associated with key.
func (m *Map[K, V]) Get(key K) (V, bool) {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}

	elem, ok := m.entries[key]
	if !ok {
		var zero V
		return zero, false
	}
	return elem.Value, true
}

// GetOrDefault retrieves the value or returns defaultVal if missing.
func (m *Map[K, V]) GetOrDefault(key K, defaultVal V) V {
	if val, ok := m.Get(key); ok {
		return val
	}
	return defaultVal
}

// Has checks if the key exists in the map.
func (m *Map[K, V]) Has(key K) bool {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}
	_, ok := m.entries[key]
	return ok
}

// Delete removes the key and its element in O(1) time.
func (m *Map[K, V]) Delete(key K) bool {
	if m.concurrent {
		m.mu.Lock()
		defer m.mu.Unlock()
	}

	elem, ok := m.entries[key]
	if !ok {
		return false
	}

	delete(m.entries, key)

	if elem.prev != nil {
		elem.prev.next = elem.next
	} else {
		m.head = elem.next
	}

	if elem.next != nil {
		elem.next.prev = elem.prev
	} else {
		m.tail = elem.prev
	}

	elem.prev = nil
	elem.next = nil
	return true
}

// Len returns the number of elements in the map.
func (m *Map[K, V]) Len() int {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}
	return len(m.entries)
}

// Clear resets the ordered map to empty.
func (m *Map[K, V]) Clear() {
	if m.concurrent {
		m.mu.Lock()
		defer m.mu.Unlock()
	}
	m.entries = make(map[K]*Element[K, V])
	m.head = nil
	m.tail = nil
}

// Front returns the first element in insertion order.
func (m *Map[K, V]) Front() *Element[K, V] {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}
	return m.head
}

// Back returns the last element in insertion order.
func (m *Map[K, V]) Back() *Element[K, V] {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}
	return m.tail
}

// Keys returns all keys in insertion order.
func (m *Map[K, V]) Keys() []K {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}

	keys := make([]K, 0, len(m.entries))
	for curr := m.head; curr != nil; curr = curr.next {
		keys = append(keys, curr.Key)
	}
	return keys
}

// Values returns all values in insertion order.
func (m *Map[K, V]) Values() []V {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}

	vals := make([]V, 0, len(m.entries))
	for curr := m.head; curr != nil; curr = curr.next {
		vals = append(vals, curr.Value)
	}
	return vals
}

// Entries returns a slice of Entry structs in insertion order.
func (m *Map[K, V]) Entries() []Entry[K, V] {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}

	entries := make([]Entry[K, V], 0, len(m.entries))
	for curr := m.head; curr != nil; curr = curr.next {
		entries = append(entries, Entry[K, V]{Key: curr.Key, Value: curr.Value})
	}
	return entries
}

// ForEach iterates through the map in insertion order. Return false from fn to break early.
func (m *Map[K, V]) ForEach(fn func(key K, value V) bool) {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}

	for curr := m.head; curr != nil; curr = curr.next {
		if !fn(curr.Key, curr.Value) {
			break
		}
	}
}

// Sort reorders the elements in-place according to the given less function.
func (m *Map[K, V]) Sort(less func(a, b Entry[K, V]) bool) {
	if m.concurrent {
		m.mu.Lock()
		defer m.mu.Unlock()
	}

	if len(m.entries) <= 1 {
		return
	}

	entries := make([]Entry[K, V], 0, len(m.entries))
	for curr := m.head; curr != nil; curr = curr.next {
		entries = append(entries, Entry[K, V]{Key: curr.Key, Value: curr.Value})
	}

	sort.Slice(entries, func(i, j int) bool {
		return less(entries[i], entries[j])
	})

	// Rebuild linked list
	m.head = nil
	m.tail = nil
	for i := range entries {
		elem := m.entries[entries[i].Key]
		elem.prev = m.tail
		elem.next = nil
		if m.tail != nil {
			m.tail.next = elem
		} else {
			m.head = elem
		}
		m.tail = elem
	}
}

// MarshalJSON serializes the ordered map as a JSON object, preserving insertion order.
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	if m.concurrent {
		m.mu.RLock()
		defer m.mu.RUnlock()
	}

	var buf bytes.Buffer
	buf.WriteByte('{')

	first := true
	for curr := m.head; curr != nil; curr = curr.next {
		if !first {
			buf.WriteByte(',')
		}
		first = false

		keyBytes, err := json.Marshal(fmt.Sprint(curr.Key))
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		valBytes, err := json.Marshal(curr.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// UnmarshalJSON deserializes a JSON object into an OrderedMap[string, V], preserving key order.
func (m *Map[string, V]) UnmarshalJSON(data []byte) error {
	if m.entries == nil {
		m.entries = make(map[string]*Element[string, V])
	}
	m.Clear()

	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return ErrInvalidJSON
	}

	for dec.More() {
		tokKey, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := tokKey.(string)
		if !ok {
			return ErrInvalidJSON
		}

		var val V
		if err := dec.Decode(&val); err != nil {
			return err
		}
		m.Set(key, val)
	}

	// consume closing '}'
	_, err = dec.Token()
	return err
}
