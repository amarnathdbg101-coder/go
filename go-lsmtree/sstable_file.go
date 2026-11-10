package lsmtree

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"io"
	"os"
	"sort"
)

type sstEntry struct {
	key     []byte
	val     []byte
	deleted bool
}

type sstBloomFilter struct {
	bits []byte
	k    uint8
}

func newSSTBloom(keys [][]byte) *sstBloomFilter {
	n := len(keys)
	if n == 0 {
		return &sstBloomFilter{bits: make([]byte, 8), k: 3}
	}
	numBits := n * 10
	if numBits < 64 {
		numBits = 64
	}
	numBytes := (numBits + 7) / 8
	bf := &sstBloomFilter{
		bits: make([]byte, numBytes),
		k:    7,
	}
	for _, k := range keys {
		h := fnv.New64a()
		_, _ = h.Write(k)
		h1 := h.Sum64()
		h2 := h1>>32 | h1<<32
		if h2%2 == 0 {
			h2++
		}
		for i := uint8(0); i < bf.k; i++ {
			bit := (h1 + uint64(i)*h2) % uint64(numBytes*8)
			bf.bits[bit/8] |= (1 << (bit % 8))
		}
	}
	return bf
}

func (b *sstBloomFilter) contains(key []byte) bool {
	if len(b.bits) == 0 {
		return false
	}
	numBits := len(b.bits) * 8
	h := fnv.New64a()
	_, _ = h.Write(key)
	h1 := h.Sum64()
	h2 := h1>>32 | h1<<32
	if h2%2 == 0 {
		h2++
	}
	for i := uint8(0); i < b.k; i++ {
		bit := (h1 + uint64(i)*h2) % uint64(numBits)
		if (b.bits[bit/8] & (1 << (bit % 8))) == 0 {
			return false
		}
	}
	return true
}

type sstFile struct {
	path        string
	bloom       *sstBloomFilter
	entries     []sstEntry
	minKey      []byte
	maxKey      []byte
}

func writeSSTable(path string, entries []entry) (*sstFile, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keys := make([][]byte, 0, len(entries))
	var sstEntries []sstEntry

	for _, e := range entries {
		var flag byte
		if e.deleted {
			flag = 1
		}

		var header [7]byte
		binary.BigEndian.PutUint16(header[0:2], uint16(len(e.key)))
		binary.BigEndian.PutUint32(header[2:6], uint32(len(e.val)))
		header[6] = flag

		if _, err := f.Write(header[:]); err != nil {
			return nil, err
		}
		if _, err := f.Write(e.key); err != nil {
			return nil, err
		}
		if len(e.val) > 0 {
			if _, err := f.Write(e.val); err != nil {
				return nil, err
			}
		}

		keys = append(keys, e.key)
		sstEntries = append(sstEntries, sstEntry{
			key:     e.key,
			val:     e.val,
			deleted: e.deleted,
		})
	}

	if err := f.Sync(); err != nil {
		return nil, err
	}

	bf := newSSTBloom(keys)
	var minK, maxK []byte
	if len(entries) > 0 {
		minK = entries[0].key
		maxK = entries[len(entries)-1].key
	}

	return &sstFile{
		path:    path,
		bloom:   bf,
		entries: sstEntries,
		minKey:  minK,
		maxKey:  maxK,
	}, nil
}

func readSSTable(path string) (*sstFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(data)
	var entries []sstEntry
	var keys [][]byte

	for {
		var header [7]byte
		_, err := io.ReadFull(r, header[:])
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return nil, err
		}

		kLen := int(binary.BigEndian.Uint16(header[0:2]))
		vLen := int(binary.BigEndian.Uint32(header[2:6]))
		deleted := header[6] == 1

		k := make([]byte, kLen)
		if _, err := io.ReadFull(r, k); err != nil {
			return nil, err
		}

		v := make([]byte, vLen)
		if vLen > 0 {
			if _, err := io.ReadFull(r, v); err != nil {
				return nil, err
			}
		}

		entries = append(entries, sstEntry{
			key:     k,
			val:     v,
			deleted: deleted,
		})
		keys = append(keys, k)
	}

	bf := newSSTBloom(keys)
	var minK, maxK []byte
	if len(entries) > 0 {
		minK = entries[0].key
		maxK = entries[len(entries)-1].key
	}

	return &sstFile{
		path:    path,
		bloom:   bf,
		entries: entries,
		minKey:  minK,
		maxKey:  maxK,
	}, nil
}

func (s *sstFile) get(key []byte) ([]byte, bool, bool) {
	// 1. Bloom Filter check (0 scan if absent)
	if !s.bloom.contains(key) {
		return nil, false, false
	}

	// 2. Range boundary check
	if len(s.minKey) > 0 && bytes.Compare(key, s.minKey) < 0 {
		return nil, false, false
	}
	if len(s.maxKey) > 0 && bytes.Compare(key, s.maxKey) > 0 {
		return nil, false, false
	}

	// 3. Binary search
	idx := sort.Search(len(s.entries), func(i int) bool {
		return bytes.Compare(s.entries[i].key, key) >= 0
	})

	if idx < len(s.entries) && bytes.Equal(s.entries[idx].key, key) {
		e := s.entries[idx]
		if e.deleted {
			return nil, true, true // found tombstone
		}
		return e.val, false, true
	}

	return nil, false, false
}
