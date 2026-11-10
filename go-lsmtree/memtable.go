package lsmtree

import (
	"bytes"
	"sort"
	"sync"
)

type entry struct {
	key     []byte
	val     []byte
	deleted bool
}

type memTable struct {
	mu      sync.RWMutex
	entries []entry
	size    int
}

func newMemTable() *memTable {
	return &memTable{
		entries: make([]entry, 0, 64),
	}
}

func (m *memTable) put(key, val []byte, deleted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := sort.Search(len(m.entries), func(i int) bool {
		return bytes.Compare(m.entries[i].key, key) >= 0
	})

	if idx < len(m.entries) && bytes.Equal(m.entries[idx].key, key) {
		// Replace existing
		m.size -= len(m.entries[idx].val)
		m.entries[idx].val = val
		m.entries[idx].deleted = deleted
		m.size += len(val)
		return
	}

	// Insert sorted
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	var valCopy []byte
	if len(val) > 0 {
		valCopy = make([]byte, len(val))
		copy(valCopy, val)
	}

	m.entries = append(m.entries, entry{})
	copy(m.entries[idx+1:], m.entries[idx:])
	m.entries[idx] = entry{
		key:     keyCopy,
		val:     valCopy,
		deleted: deleted,
	}
	m.size += len(keyCopy) + len(valCopy)
}

func (m *memTable) get(key []byte) ([]byte, bool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	idx := sort.Search(len(m.entries), func(i int) bool {
		return bytes.Compare(m.entries[i].key, key) >= 0
	})

	if idx < len(m.entries) && bytes.Equal(m.entries[idx].key, key) {
		e := m.entries[idx]
		if e.deleted {
			return nil, true, true // found, but marked deleted (tombstone)
		}
		valCopy := make([]byte, len(e.val))
		copy(valCopy, e.val)
		return valCopy, false, true
	}

	return nil, false, false // not in memtable
}

func (m *memTable) allEntries() []entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cp := make([]entry, len(m.entries))
	copy(cp, m.entries)
	return cp
}

func (m *memTable) byteSize() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size
}

func (m *memTable) len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}
