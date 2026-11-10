# Specification: Log-Structured Merge Tree (`go-lsmtree`)

## 1. LSM Architecture & Data Flow

```
[Client Put/Delete]
         │
         ├──► 1. WAL (Append-Only Sequential Disk I/O)
         │
         └──► 2. Active MemTable (In-Memory Sorted Buffer)
                     │
                     ▼ (Capacity Threshold Reached)
              3. Immutable SSTable Flush (`000001.sst`)
                     │
                     ▼
              4. Compaction (Multi-way merge, purge tombstones & stale versions)
```

## 2. Invariants & Lookup Path
1. **Read Path Hierarchy**:
   $$\text{Active MemTable} \longrightarrow \text{SSTable}_N (\text{Newest}) \longrightarrow \dots \longrightarrow \text{SSTable}_0 (\text{Oldest})$$
   First match returns immediately. If match is marked `deleted == true`, returns `ErrKeyNotFound`.
2. **Durability Invariant**:
   If `SyncWAL == true`, `Put` or `Delete` returns success only after the OS commits the entry to physical storage via `os.File.Sync()`.
3. **Compaction Invariant**:
   During compaction of $\{S_1, S_2, \dots, S_k\}$, for each unique key $K$, the version from the newest SSTable is retained. If that version is a tombstone, it is completely dropped from the output SSTable.
