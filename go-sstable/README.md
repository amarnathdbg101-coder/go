# go-sstable

`go-sstable` provides an immutable, high-efficiency Sorted String Table (SSTable) file storage engine in pure Go. It is the core on-disk building block of modern Log-Structured Merge-tree (LSM) databases such as LevelDB, RocksDB, Cassandra, ScyllaDB, and Google Bigtable.

## Key Features

- **Immutable On-Disk Data Blocks**: Fixed-size blocks (e.g. 4KB) containing lexicographically sorted binary key-value entries.
- **Embedded Bloom Filter Pre-Filtering**: Fast $O(1)$ in-memory bloom filter lookup rejects non-existent keys without executing any physical disk I/O.
- **Binary Search Index Block**: Direct pointer navigation jumps straight to the exact block containing the queried key in $O(\log B)$ time.
- **Bidirectional Range Scans**: Fast sequential iteration over arbitrary key ranges `[startKey, endKey]`.
- **Strict Monotonic Ordering**: Guaranteed sorted invariants enforced at writer append time.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-sstable"
)

func main() {
	path := "storage_001.sst"

	// 1. Write sorted entries
	w, err := sstable.NewWriter(path, sstable.WithBlockSize(4096))
	if err != nil {
		panic(err)
	}

	_ = w.Append([]byte("user:1001"), []byte("Amar"))
	_ = w.Append([]byte("user:1002"), []byte("Rahul"))
	_ = w.Append([]byte("user:1003"), []byte("Priya"))
	_ = w.Close()

	// 2. Open Reader for sub-millisecond point lookups
	r, err := sstable.Open(path)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	val, err := r.Get([]byte("user:1002"))
	if err == nil {
		fmt.Printf("Found user: %s\n", string(val)) // Rahul
	}

	// 3. Range Scan
	_ = r.Scan([]byte("user:1001"), []byte("user:1003"), func(k, v []byte) error {
		fmt.Printf("%s -> %s\n", string(k), string(v))
		return nil
	})
}
```
