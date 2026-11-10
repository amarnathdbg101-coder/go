package sstable

import (
	"bytes"
	"encoding/binary"
	"os"
	"sort"
	"sync"
)

const (
	magicNumber = uint64(0x53535441424C4531) // "SSTABLE1"
	footerSize  = 48
)

type indexEntry struct {
	lastKey []byte
	offset  uint64
	size    uint32
}

// -------------------------------------------------------------
// SSTable Writer
// -------------------------------------------------------------

// Writer constructs an immutable SSTable file.
type Writer struct {
	cfg        Config
	file       *os.File
	offset     uint64
	currentBuf bytes.Buffer
	lastKey    []byte
	prevKey    []byte
	index      []indexEntry
	allKeys    [][]byte
	closed     bool
}

// NewWriter creates a new SSTable writer targeting the specified file path.
func NewWriter(path string, opts ...Option) (*Writer, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Writer{
		cfg:  cfg,
		file: f,
	}, nil
}

// Append inserts a key-value pair. Keys must be added in strictly ascending order.
func (w *Writer) Append(key, val []byte) error {
	if w.closed {
		return ErrWriterClosed
	}
	if len(w.prevKey) > 0 && bytes.Compare(key, w.prevKey) <= 0 {
		return ErrUnsortedKeys
	}

	// Append to current block
	var lenBuf [6]byte
	binary.BigEndian.PutUint16(lenBuf[0:2], uint16(len(key)))
	binary.BigEndian.PutUint32(lenBuf[2:6], uint32(len(val)))

	w.currentBuf.Write(lenBuf[:])
	w.currentBuf.Write(key)
	w.currentBuf.Write(val)

	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	w.allKeys = append(w.allKeys, keyCopy)
	w.prevKey = keyCopy
	w.lastKey = keyCopy

	// Flush block if threshold exceeded
	if w.currentBuf.Len() >= w.cfg.BlockSize {
		if err := w.flushBlock(); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) flushBlock() error {
	if w.currentBuf.Len() == 0 {
		return nil
	}

	blockSize := uint32(w.currentBuf.Len())
	blockOffset := w.offset

	if _, err := w.file.Write(w.currentBuf.Bytes()); err != nil {
		return err
	}

	w.offset += uint64(blockSize)
	w.index = append(w.index, indexEntry{
		lastKey: w.lastKey,
		offset:  blockOffset,
		size:    blockSize,
	})

	w.currentBuf.Reset()
	return nil
}

// Close finishes writing blocks, bloom filter, index block, and footer.
func (w *Writer) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true

	// Flush remaining entries
	if err := w.flushBlock(); err != nil {
		_ = w.file.Close()
		return err
	}

	// 1. Write Bloom Filter Block
	bf := newBloomFilter(w.allKeys, w.cfg.BloomBitsPerKey)
	bloomOffset := w.offset
	var bloomPayload bytes.Buffer
	bloomPayload.WriteByte(bf.k)
	bloomPayload.Write(bf.bits)

	if _, err := w.file.Write(bloomPayload.Bytes()); err != nil {
		_ = w.file.Close()
		return err
	}
	bloomSize := uint64(bloomPayload.Len())
	w.offset += bloomSize

	// 2. Write Index Block
	indexOffset := w.offset
	var indexBuf bytes.Buffer
	for _, entry := range w.index {
		var entryHeader [14]byte
		binary.BigEndian.PutUint16(entryHeader[0:2], uint16(len(entry.lastKey)))
		binary.BigEndian.PutUint64(entryHeader[2:10], entry.offset)
		binary.BigEndian.PutUint32(entryHeader[10:14], entry.size)

		indexBuf.Write(entryHeader[:])
		indexBuf.Write(entry.lastKey)
	}

	if _, err := w.file.Write(indexBuf.Bytes()); err != nil {
		_ = w.file.Close()
		return err
	}
	indexSize := uint64(indexBuf.Len())
	w.offset += indexSize

	// 3. Write Footer (48 bytes)
	var footer [footerSize]byte
	binary.BigEndian.PutUint64(footer[0:8], magicNumber)
	binary.BigEndian.PutUint64(footer[8:16], indexOffset)
	binary.BigEndian.PutUint64(footer[16:24], indexSize)
	binary.BigEndian.PutUint64(footer[24:32], bloomOffset)
	binary.BigEndian.PutUint64(footer[32:40], bloomSize)
	binary.BigEndian.PutUint64(footer[40:48], uint64(len(w.allKeys)))

	if _, err := w.file.Write(footer[:]); err != nil {
		_ = w.file.Close()
		return err
	}

	if err := w.file.Sync(); err != nil {
		_ = w.file.Close()
		return err
	}
	return w.file.Close()
}

// -------------------------------------------------------------
// SSTable Reader
// -------------------------------------------------------------

// Reader provides random-access lookups and sequential range scans on an immutable SSTable file.
type Reader struct {
	file        *os.File
	bloomFilter *bloomFilter
	index       []indexEntry
	totalKeys   uint64
	mu          sync.RWMutex
}

// Open reads the SSTable footer, bloom filter, and index block into memory.
func Open(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if fi.Size() < footerSize {
		_ = f.Close()
		return nil, ErrCorruptFooter
	}

	// 1. Read Footer
	var footer [footerSize]byte
	if _, err := f.ReadAt(footer[:], fi.Size()-footerSize); err != nil {
		_ = f.Close()
		return nil, err
	}

	magic := binary.BigEndian.Uint64(footer[0:8])
	if magic != magicNumber {
		_ = f.Close()
		return nil, ErrCorruptFooter
	}

	indexOffset := int64(binary.BigEndian.Uint64(footer[8:16]))
	indexSize := int(binary.BigEndian.Uint64(footer[16:24]))
	bloomOffset := int64(binary.BigEndian.Uint64(footer[24:32]))
	bloomSize := int(binary.BigEndian.Uint64(footer[32:40]))
	totalKeys := binary.BigEndian.Uint64(footer[40:48])

	// 2. Read Bloom Filter
	bloomData := make([]byte, bloomSize)
	if _, err := f.ReadAt(bloomData, bloomOffset); err != nil {
		_ = f.Close()
		return nil, err
	}
	bf := &bloomFilter{
		k:    bloomData[0],
		bits: bloomData[1:],
	}

	// 3. Read Index Block
	indexData := make([]byte, indexSize)
	if _, err := f.ReadAt(indexData, indexOffset); err != nil {
		_ = f.Close()
		return nil, err
	}

	var index []indexEntry
	offset := 0
	for offset < len(indexData) {
		keyLen := int(binary.BigEndian.Uint16(indexData[offset : offset+2]))
		blockOffset := binary.BigEndian.Uint64(indexData[offset+2 : offset+10])
		blockSize := binary.BigEndian.Uint32(indexData[offset+10 : offset+14])
		offset += 14

		key := indexData[offset : offset+keyLen]
		offset += keyLen

		entryKey := make([]byte, keyLen)
		copy(entryKey, key)

		index = append(index, indexEntry{
			lastKey: entryKey,
			offset:  blockOffset,
			size:    blockSize,
		})
	}

	return &Reader{
		file:        f,
		bloomFilter: bf,
		index:       index,
		totalKeys:   totalKeys,
	}, nil
}

// Get performs a point lookup for a key. Returns ErrKeyNotFound if key does not exist.
func (r *Reader) Get(key []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Fast Bloom Filter pre-check (eliminates disk read on misses)
	if !r.bloomFilter.contains(key) {
		return nil, ErrKeyNotFound
	}

	// 2. Binary search on index to find candidate block
	idx := sort.Search(len(r.index), func(i int) bool {
		return bytes.Compare(r.index[i].lastKey, key) >= 0
	})
	if idx >= len(r.index) {
		return nil, ErrKeyNotFound
	}

	// 3. Read and scan block
	candidate := r.index[idx]
	blockData := make([]byte, candidate.size)
	if _, err := r.file.ReadAt(blockData, int64(candidate.offset)); err != nil {
		return nil, err
	}

	pos := 0
	for pos < len(blockData) {
		kLen := int(binary.BigEndian.Uint16(blockData[pos : pos+2]))
		vLen := int(binary.BigEndian.Uint32(blockData[pos+2 : pos+6]))
		pos += 6

		k := blockData[pos : pos+kLen]
		pos += kLen
		v := blockData[pos : pos+vLen]
		pos += vLen

		if bytes.Equal(k, key) {
			valCopy := make([]byte, vLen)
			copy(valCopy, v)
			return valCopy, nil
		}
		if bytes.Compare(k, key) > 0 {
			break
		}
	}

	return nil, ErrKeyNotFound
}

// Scan performs a range scan over [startKey, endKey] in ascending order.
func (r *Reader) Scan(startKey, endKey []byte, fn func(key, val []byte) error) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	startBlockIdx := 0
	if len(startKey) > 0 {
		startBlockIdx = sort.Search(len(r.index), func(i int) bool {
			return bytes.Compare(r.index[i].lastKey, startKey) >= 0
		})
		if startBlockIdx >= len(r.index) {
			return nil // out of range
		}
	}

	for i := startBlockIdx; i < len(r.index); i++ {
		candidate := r.index[i]
		blockData := make([]byte, candidate.size)
		if _, err := r.file.ReadAt(blockData, int64(candidate.offset)); err != nil {
			return err
		}

		pos := 0
		for pos < len(blockData) {
			kLen := int(binary.BigEndian.Uint16(blockData[pos : pos+2]))
			vLen := int(binary.BigEndian.Uint32(blockData[pos+2 : pos+6]))
			pos += 6

			k := blockData[pos : pos+kLen]
			pos += kLen
			v := blockData[pos : pos+vLen]
			pos += vLen

			if len(startKey) > 0 && bytes.Compare(k, startKey) < 0 {
				continue
			}
			if len(endKey) > 0 && bytes.Compare(k, endKey) > 0 {
				return nil
			}

			if err := fn(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// Count returns the total number of keys in the SSTable.
func (r *Reader) Count() uint64 {
	return r.totalKeys
}

// Close closes the underlying SSTable file descriptor.
func (r *Reader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Close()
}
