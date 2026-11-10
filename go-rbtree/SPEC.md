# Specification: Left-Leaning Red-Black Tree (`go-rbtree`)

## 1. LLRB Structural Invariants (Sedgewick, 2008)
1. **Left-Leaning Property**: Red links are strictly forbidden on the right branch of any node:
   $$\text{isRed}(h.\text{right}) \implies \text{false}$$
2. **No Consecutive Red Links**:
   $$\text{isRed}(h) \implies \neg\text{isRed}(h.\text{left})$$
3. **Black Balance**: Every path from the root to any null link contains the exact same number of black links.
4. **Height Bound**:
   $$h \le 2 \log_2(N + 1)$$

## 2. Complexity Guarantees
- `Get(key)`: $O(\log N)$ worst-case.
- `Insert(key, val)`: $O(\log N)$ worst-case.
- `Floor(key)` / `Ceiling(key)`: $O(\log N)$.
- `AscendRange(start, end)`: $O(\log N + K)$ where $K$ is the number of keys in range.
