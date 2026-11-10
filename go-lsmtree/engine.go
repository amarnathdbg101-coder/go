package lsmtree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	walFileName = "wal.log"
	sstExt      = ".sst"
)

// DB represents a Log-Structured Merge Tree key-value database.
type DB struct {
	mu         sync.RWMutex
	cfg        Config
	activeMem  *memTable
	walFile    *os.File
	sstables   []*sstFile // newest first
	nextSSTSeq int
	closed     bool
}

// Open creates or recovers an LSM-Tree database instance.
func Open(opts ...Option) (*DB, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, err
	}

	db := &DB{
		cfg:       cfg,
		activeMem: newMemTable(),
	}

	// 1. Discover existing SSTables
	if err := db.loadSSTables(); err != nil {
		return nil, err
	}

	// 2. Replay WAL if present
	walPath := filepath.Join(cfg.Dir, walFileName)
	if err := db.replayWAL(walPath); err != nil {
		return nil, err
	}

	// 3. Open WAL for appending
	wf, err := os.OpenFile(walPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	db.walFile = wf

	return db, nil
}

// Put writes or overwrites a key-value pair.
func (db *DB) Put(key, val []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return ErrDBClosed
	}

	// Write to WAL
	if err := db.writeWALEntry(key, val, false); err != nil {
		return err
	}

	// Update MemTable
	db.activeMem.put(key, val, false)

	// Flush if threshold reached
	if db.activeMem.byteSize() >= db.cfg.MemTableCapacity {
		if err := db.flushMemTableInternal(); err != nil {
			return err
		}
	}
	return nil
}

// Delete marks a key as deleted (tombstone).
func (db *DB) Delete(key []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return ErrDBClosed
	}

	// Write tombstone to WAL
	if err := db.writeWALEntry(key, nil, true); err != nil {
		return err
	}

	// Mark tombstone in MemTable
	db.activeMem.put(key, nil, true)

	if db.activeMem.byteSize() >= db.cfg.MemTableCapacity {
		if err := db.flushMemTableInternal(); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves the value associated with key. Returns ErrKeyNotFound if missing or deleted.
func (db *DB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrEmptyKey
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, ErrDBClosed
	}

	// 1. Check active MemTable
	val, deleted, found := db.activeMem.get(key)
	if found {
		if deleted {
			return nil, ErrKeyNotFound
		}
		return val, nil
	}

	// 2. Check SSTables from newest to oldest
	for _, sst := range db.sstables {
		val, deleted, found := sst.get(key)
		if found {
			if deleted {
				return nil, ErrKeyNotFound
			}
			valCopy := make([]byte, len(val))
			copy(valCopy, val)
			return valCopy, nil
		}
	}

	return nil, ErrKeyNotFound
}

// Flush explicitly flushes the active MemTable to an immutable SSTable file.
func (db *DB) Flush() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return ErrDBClosed
	}
	return db.flushMemTableInternal()
}

func (db *DB) flushMemTableInternal() error {
	if db.activeMem.len() == 0 {
		return nil
	}

	db.nextSSTSeq++
	fileName := fmt.Sprintf("%06d%s", db.nextSSTSeq, sstExt)
	filePath := filepath.Join(db.cfg.Dir, fileName)

	entries := db.activeMem.allEntries()
	sst, err := writeSSTable(filePath, entries)
	if err != nil {
		return err
	}

	// Prepend new SSTable (newest first)
	db.sstables = append([]*sstFile{sst}, db.sstables...)

	// Reset MemTable
	db.activeMem = newMemTable()

	// Truncate WAL
	_ = db.walFile.Close()
	walPath := filepath.Join(db.cfg.Dir, walFileName)
	wf, err := os.OpenFile(walPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	db.walFile = wf
	return nil
}

// Compact merges all SSTables into a single consolidated SSTable, purging tombstones and stale versions.
func (db *DB) Compact() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return ErrDBClosed
	}

	if len(db.sstables) <= 1 {
		return nil // Nothing to compact
	}

	// Collect newest version of each key across all SSTables (oldest to newest iteration)
	consolidated := make(map[string]entry)
	for i := len(db.sstables) - 1; i >= 0; i-- {
		sst := db.sstables[i]
		for _, e := range sst.entries {
			consolidated[string(e.key)] = entry{
				key:     e.key,
				val:     e.val,
				deleted: e.deleted,
			}
		}
	}

	// Filter out tombstones and sort
	var merged []entry
	for _, e := range consolidated {
		if !e.deleted {
			merged = append(merged, e)
		}
	}

	sort.Slice(merged, func(i, j int) bool {
		return bytes.Compare(merged[i].key, merged[j].key) < 0
	})

	// Write new compacted SSTable
	db.nextSSTSeq++
	compactedName := fmt.Sprintf("%06d%s", db.nextSSTSeq, sstExt)
	compactedPath := filepath.Join(db.cfg.Dir, compactedName)

	newSST, err := writeSSTable(compactedPath, merged)
	if err != nil {
		return err
	}

	// Delete old SSTable files
	for _, oldSST := range db.sstables {
		_ = os.Remove(oldSST.path)
	}

	db.sstables = []*sstFile{newSST}
	return nil
}

// Scan iterates over all active keys within range [startKey, endKey] in ascending order.
func (db *DB) Scan(startKey, endKey []byte, fn func(k, v []byte) error) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return ErrDBClosed
	}

	// Build consolidated snapshot
	snapshot := make(map[string]entry)

	// SSTables from oldest to newest
	for i := len(db.sstables) - 1; i >= 0; i-- {
		for _, e := range db.sstables[i].entries {
			snapshot[string(e.key)] = entry{
				key:     e.key,
				val:     e.val,
				deleted: e.deleted,
			}
		}
	}

	// Overlay MemTable
	for _, e := range db.activeMem.allEntries() {
		snapshot[string(e.key)] = e
	}

	// Filter and sort keys
	var sortedKeys []string
	for k, e := range snapshot {
		if e.deleted {
			continue
		}
		keyBytes := []byte(k)
		if len(startKey) > 0 && bytes.Compare(keyBytes, startKey) < 0 {
			continue
		}
		if len(endKey) > 0 && bytes.Compare(keyBytes, endKey) > 0 {
			continue
		}
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		e := snapshot[k]
		if err := fn(e.key, e.val); err != nil {
			return err
		}
	}
	return nil
}

// Close flushes active state and closes open file descriptors.
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return nil
	}
	db.closed = true

	if db.walFile != nil {
		_ = db.walFile.Sync()
		_ = db.walFile.Close()
	}
	return nil
}

// -------------------------------------------------------------
// Internal WAL & SSTable Helpers
// -------------------------------------------------------------

func (db *DB) writeWALEntry(key, val []byte, deleted bool) error {
	var flag byte
	if deleted {
		flag = 1
	}

	var header [7]byte
	binary.BigEndian.PutUint16(header[0:2], uint16(len(key)))
	binary.BigEndian.PutUint32(header[2:6], uint32(len(val)))
	header[6] = flag

	if _, err := db.walFile.Write(header[:]); err != nil {
		return err
	}
	if _, err := db.walFile.Write(key); err != nil {
		return err
	}
	if len(val) > 0 {
		if _, err := db.walFile.Write(val); err != nil {
			return err
		}
	}

	if db.cfg.SyncWAL {
		return db.walFile.Sync()
	}
	return nil
}

func (db *DB) replayWAL(path string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	r := bytes.NewReader(data)
	for {
		var header [7]byte
		_, err := io.ReadFull(r, header[:])
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return err
		}

		kLen := int(binary.BigEndian.Uint16(header[0:2]))
		vLen := int(binary.BigEndian.Uint32(header[2:6]))
		deleted := header[6] == 1

		k := make([]byte, kLen)
		if _, err := io.ReadFull(r, k); err != nil {
			break
		}

		v := make([]byte, vLen)
		if vLen > 0 {
			if _, err := io.ReadFull(r, v); err != nil {
				break
			}
		}

		db.activeMem.put(k, v, deleted)
	}
	return nil
}

func (db *DB) loadSSTables() error {
	entries, err := os.ReadDir(db.cfg.Dir)
	if err != nil {
		return err
	}

	var sstFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), sstExt) {
			sstFiles = append(sstFiles, entry.Name())
		}
	}

	sort.Strings(sstFiles) // Ascending sequence: 000001.sst, 000002.sst

	for _, name := range sstFiles {
		base := strings.TrimSuffix(name, sstExt)
		seq, err := strconv.Atoi(base)
		if err == nil && seq > db.nextSSTSeq {
			db.nextSSTSeq = seq
		}

		sst, err := readSSTable(filepath.Join(db.cfg.Dir, name))
		if err != nil {
			return err
		}
		// Prepend so newest is first
		db.sstables = append([]*sstFile{sst}, db.sstables...)
	}

	return nil
}
