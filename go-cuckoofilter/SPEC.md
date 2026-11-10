# Specification: Cuckoo Filter (`go-cuckoofilter`)

## 1. Mathematical Architecture
A Cuckoo Filter organizes an array of $m$ buckets, where each bucket holds $b=4$ fingerprint slots of size $f=16$ bits ($2$ bytes).

For an item $x$:
- Fingerprint: $f_p = \text{hash}(x) \pmod{2^{16}}$ (non-zero).
- Primary bucket index: $i_1 = \text{hash}(x) \pmod m$.
- Alternate bucket index:
  $$i_2 = (i_1 \oplus \text{hash}(f_p)) \pmod m$$
Because $a \oplus b \oplus b = a$, the primary index can be derived symmetrically from $i_2$:
$$i_1 = (i_2 \oplus \text{hash}(f_p)) \pmod m$$
This property allows displacing an item without storing or recalculating the original key $x$.

## 2. Invariants & Complexity
- `Lookup(x)`: Checks $2 \times b = 8$ slots total $\implies O(1)$ worst-case.
- `Delete(x)`: Checks $2 \times b = 8$ slots total, clears matching slot $\implies O(1)$ worst-case.
- `Insert(x)`:
  - If $i_1$ or $i_2$ has an empty slot: $O(1)$.
  - If both buckets are full: Cuckoo kicks up to $\text{MaxKicks} = 500$ relocations. Amortized $O(1)$.
- **Theoretical False Positive Probability**:
  $$\epsilon \approx 2b / 2^f = \frac{2 \times 4}{2^{16}} = \frac{8}{65536} \approx 0.012\%$$
