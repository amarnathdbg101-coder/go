# Specification: Compact Radix Tree (`go-radix`)

## 1. Architectural Invariants
1. **Edge Non-Emptiness**: Every edge between nodes has a non-empty string label ($\text{len}(\text{label}) \ge 1$).
2. **Branching Constraint**: For any internal node with no value ($\text{hasValue} = \text{false}$), it must have at least 2 child edges ($\text{len}(\text{edges}) \ge 2$). If a deletion leaves an unvalued node with 1 child, it is merged with its child.
3. **Prefix Disjointness**: No two outgoing edges from the same node share the same initial byte character.

## 2. Operations Complexity
Let $K$ be the length of the queried string key:
- `Insert(key, val)`: $O(K)$.
- `Get(key)`: $O(K)$ worst case, typically terminates early in $O(1)$ on mismatch.
- `LongestPrefix(query)`: $O(K)$.
- `Delete(key)`: $O(K)$ with $O(1)$ edge reclamation.
- Space Complexity: $O(N)$ where $N$ is the number of keys, independent of redundant shared prefixes.
