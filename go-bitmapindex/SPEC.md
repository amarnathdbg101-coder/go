# Specification: Bit-Sliced Indexing (`go-bitmapindex`)

## 1. Columnar Representation
For a column of unsigned integers with bit width $B$ ($1 \le B \le 64$):
Each value $v$ is decomposed into binary digits:
$$v = \sum_{b=0}^{B-1} 2^b \cdot \text{bit}(v, b)$$

The index maintains:
- $B$ separate bitsets $S_0, S_1, \dots, S_{B-1}$, where $S_b$ contains the $b$-th bit of every row.
- An existence bitset $E$ indicating row population.

## 2. Bitwise Aggregation Invariant
The exact mathematical sum across a subset of rows $M$ is:
$$\text{SUM}(M) = \sum_{b=0}^{B-1} 2^b \cdot \text{POPCNT}(S_b \cap M)$$
- **Complexity**: $O(B)$ bitwise operations instead of $O(|M|)$ raw row fetches.
- **Hardware Acceleration**: Each word uses CPU `POPCNT` instructions (`bits.OnesCount64`).
