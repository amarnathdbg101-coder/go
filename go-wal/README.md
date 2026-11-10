# go-wal

`go-wal` provides a production-grade, crash-resilient Write-Ahead Log (WAL) engine in pure Go. It guarantees durability and zero data loss for storage engines, embedded databases, event brokers, and Raft consensus algorithms by sequentially appending mutation records to disk with CRC32 checksums before memory application.

## Key Features

- **Segmented Append-Only Architecture**: Automatically rolls log files (`00000001.wal`, `00000002.wal`) when segment byte thresholds are reached.
- **Hardware-Accelerated CRC32-IEEE Verification**: Corrupt frames or partial writes caused by abrupt server crashes are immediately detected and rejected.
- **Monotonic Log Sequence Numbers (LSN)**: Strict sequential indexing ensures deterministic replay ordering.
- **Configurable Durability Modes**:
  - `SyncAlways`: Immediate `fsync` per write for absolute transactional durability.
  - `SyncInterval`: Batch OS buffering with periodic background `fsync` for ultra-high ingestion throughput.
- **Log Compaction & Truncation**: Safely prune rotated segments (`TruncateBefore`) after checkpointing state.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-wal"
)

func main() {
	// Open or recover WAL directory
	w, err := wal.Open(
		wal.WithDir("./data/wal"),
		wal.WithMaxSegmentSize(64*1024*1024), // 64 MB per segment
		wal.WithSyncMode(wal.SyncAlways),
	)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	// Append mutation event
	lsn, err := w.Append([]byte(`{"type":"create_invoice","id":1001,"amount":4500}`))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Logged mutation with LSN: %d\n", lsn)

	// Crash Recovery Replay
	fmt.Println("Replaying mutations...")
	err = w.Replay(1, func(rec wal.Record) error {
		fmt.Printf("Replayed LSN %d: %s\n", rec.LSN, string(rec.Data))
		return nil
	})
	if err != nil {
		panic(err)
	}
}
```
