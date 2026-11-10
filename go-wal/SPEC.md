# Specification: Crash-Resilient Write-Ahead Log (`go-wal`)

## 1. Frame Wire Format
Each record is serialized with a 24-byte header followed by variable-length payload bytes:

```
+---------------+----------------+-----------------------+--------------------+----------------+
| CRC32 (4B)    | LSN (8B)       | Timestamp (8B)        | Length (4B)        | Payload (NB)   |
+---------------+----------------+-----------------------+--------------------+----------------+
```

- **CRC32 (4 bytes)**: IEEE 802.3 polynomial calculated over `[LSN + Timestamp + Length + Payload]`.
- **LSN (8 bytes)**: Unsigned 64-bit integer, strictly monotonically increasing ($1, 2, 3, \dots$).
- **Timestamp (8 bytes)**: Unix timestamp in nanoseconds (`time.Now().UnixNano()`).
- **Length (4 bytes)**: Unsigned 32-bit payload length in bytes.
- **Payload ($N$ bytes)**: Raw application data bytes.

## 2. Invariants & Guarantees
1. **Append-Only Sequential I/O**: Files are opened with `O_APPEND`. No random writes occur on active segments.
2. **Durability Invariant**: With `SyncAlways`, no `Append()` call returns success to the caller until the operating system commits write buffers to physical persistent storage via `os.File.Sync()`.
3. **Recovery Guarantee**: Replay halts or returns `ErrCorruptRecord` if any frame fails CRC verification, preventing torn or partially written pages from poisoning application state.
4. **Segment Roll Invariant**: A segment file is closed and a new segment is allocated when $\text{CurrentSize} + \text{RecordSize} > \text{MaxSegmentSize}$.
