# Specification: Space-Saving Heavy Hitters (`go-heavyhitters`)

## 1. Mathematical Guarantees (Metwally et al.)
For stream of $N$ items and capacity $M$:
1. **False Negative Guarantee**:
   Any item $e$ whose true frequency $f_e > \frac{N}{M}$ is **guaranteed** to be present in the monitored table.
2. **Frequency Bound**:
   For any monitored item $e$ with estimated count $c_e$ and error $\epsilon_e$:
   $$c_e - \epsilon_e \le f_e \le c_e$$
3. **Maximum Error**:
   $$\epsilon_e \le \frac{N}{M}$$

## 2. Invariants
- Total recorded events strictly equals $\sum f_i$.
- Table size never exceeds `Capacity`.
