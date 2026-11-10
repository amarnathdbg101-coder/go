# go-lsmtree

`go-lsmtree` provides a production-grade, crash-resilient Log-Structured Merge-tree (LSM-Tree) embedded key-value storage engine in pure Go. It delivers ultra-high write throughput, point queries, range scans, and storage reclamation via Leveled Compaction, inspired by RocksDB, LevelDB, and BadgerDB.

## Key Features

- **Sequential Append-Only Writes**: All mutations are appended to disk sequentially via a Write-Ahead Log (WAL) before memory insertion, turning random disk writes into high-speed sequential I/O.
- **In-Memory MemTable**: Fast in-memory sorted buffer servicing real-time reads and writes.
- **Immutable SSTable Flush**: Automatically flushes MemTable to disk when memory capacity threshold is reached.
- **Tombstone Deletion & Merged Compaction**: Deletions write light tombstones; background/manual compaction consolidates multiple SSTables into a single file, eliminating stale versions and reclaiming disk space.
- **Crash Durability & Zero Data Loss**: Automatically replays uncommitted WAL records on database startup.
- **Range Scans**: Sub-millisecond sorted range queries across MemTable and on-disk SSTables.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-lsmtree"
)

func main() {
	// Open or recover LSM database
	db, err := lsmtree.Open(
		lsmtree.WithDir("./data/lsm"),
		lsmtree.WithMemTableCapacity(4*1024*1024), // 4 MB MemTable
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// High-speed write
	_ = db.Put([]byte("user:101"), []byte(`{"name":"Amar","city":"Delhi"}`))

	// Point Lookup
	val, err := db.Get([]byte("user:101"))
	if err == nil {
		fmt.Printf("User: %s\n", string(val))
	}

	// Range Scan [user:100 .. user:200]
	_ = db.Scan([]byte("user:100"), []byte("user:200"), func(k, v []byte) error {
		fmt.Printf("%s: %s\n", string(k), string(v))
		return nil
	})

	// Consolidate SSTable files and reclaim disk space
	_ = db.Compact()
}
```
