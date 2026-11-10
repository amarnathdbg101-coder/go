# Specification: Compressed Roaring Bitmap (`go-roaring`)

## 1. Mathematical Background
Roaring bitmaps partition a 32-bit unsigned integer space into chunks sharing the same 16 most-significant bits:
$$\text{integer} = \underbrace{high}_{16\text{ bits}} \mid \underbrace{low}_{16\text{ bits}}$$

Each 16-bit key maps to a container holding the lower 16 bits ($[0, 65535]$).

## 2. Container Invariants
- **Threshold**: $N = 4,096$.
  - An array of $4,096$ 16-bit unsigned integers takes $4,096 \times 2 = 8,192\text{ bytes} = 8\text{ KB}$.
  - A fixed bitset for $65,536$ bits takes $65,536 / 8 = 8,192\text{ bytes} = 8\text{ KB}$.
  - Therefore:
    - If $|S| < 4,096$: Array container is strictly more compact.
    - If $|S| \ge 4,096$: Bitmap container provides $O(1)$ lookup and bitwise SIMD/POPCNT parallelism without exceeding $8\text{ KB}$.

## 3. Operations Complexity
- `Add(x)`: $O(\log k)$ for Array container (binary search insertion), $O(1)$ for Bitmap container.
- `Contains(x)`: $O(\log k)$ for Array container, $O(1)$ for Bitmap container.
- `Cardinality()`: $O(1)$ cached/computed via `bits.OnesCount64`.
- `And(other)`: $O(N)$ linear two-pointer merge or $1,024$ CPU word bitwise AND instructions.
- `Or(other)`: $O(N)$ linear two-pointer merge or $1,024$ CPU word bitwise OR instructions.
