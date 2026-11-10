# Specification: Sorted String Table (`go-sstable`)

## 1. File Layout Architecture
An SSTable file consists of sequential Data Blocks, followed by a Bloom Filter Block, an Index Block, and a trailing 48-byte fixed-size Footer:

```
+-----------------------------------------------------------+
| Data Block 0 (Keys [k_0 .. k_m])                          |
+-----------------------------------------------------------+
| Data Block 1 (Keys [k_{m+1} .. k_n])                      |
+-----------------------------------------------------------+
| ...                                                       |
+-----------------------------------------------------------+
| Bloom Filter Block (k: 1B, bitset: NB)                    |
+-----------------------------------------------------------+
| Index Block ([LastKeyLen, Offset, Size, LastKey] ...)     |
+-----------------------------------------------------------+
| Footer (48 Bytes Fixed)                                   |
|   - Magic Number (8B: 0x53535441424C4531)                 |
|   - IndexBlockOffset (8B)                                 |
|   - IndexBlockSize (8B)                                   |
|   - BloomOffset (8B)                                      |
|   - BloomSize (8B)                                        |
|   - TotalKeysCount (8B)                                   |
+-----------------------------------------------------------+
```

## 2. Invariants & Lookups
- **Append Invariant**: $\forall i > 0, \text{Key}_i > \text{Key}_{i-1}$ under standard lexicographical byte comparison (`bytes.Compare`).
- **Point Lookup Complexity**:
  1. $\text{BloomFilter.Contains}(\text{Key})$: $O(k)$ bit tests ($k \approx 7$). Returns negative with $0$ disk I/O.
  2. $\text{BinarySearch}(\text{IndexBlock})$: $O(\log B)$ where $B = \lceil N / \text{BlockEntries} \rceil$.
  3. Single file `ReadAt` of block data ($4\text{ KB}$ typical), followed by linear scan.
- **Range Scan**: $O(\log B + K)$ where $K$ is the number of keys traversed.
