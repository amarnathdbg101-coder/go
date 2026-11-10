# Specification: Generic In-Memory B-Tree (`go-btree`)

## 1. Structural Invariants
For minimum degree $t \ge 2$:
1. **Root**: Has at least 1 key if non-empty, and at most $2t - 1$ keys.
2. **Internal Nodes**: Every internal node except root has at least $t - 1$ keys and at most $2t - 1$ keys, and between $t$ and $2t$ children.
3. **Leaves**: All leaf nodes appear at the exact same depth (balanced height).
4. **Ordering**:
   - For an internal node with keys $k_1 < k_2 < \dots < k_x$, child $c_i$ holds keys in range $(k_{i-1}, k_i)$.

## 2. Complexity Guarantees
- **Height**:
  $$h \le \log_t \left(\frac{N + 1}{2}\right)$$
  For $N = 1,000,000$ and $t = 32$, $h \le 4$.
- **Lookup (`Get`)**: $O(h \cdot \log t) = O(\log N)$.
- **Insert (`Set`)**: $O(h \cdot t)$ with contiguous slice shifts; $h \le 4$ keeps memory overhead negligible.
- **Range Iteration (`AscendRange`)**: $O(\log N + K)$ where $K$ is the number of keys traversed.
