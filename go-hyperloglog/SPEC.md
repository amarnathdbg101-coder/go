# Specification: HyperLogLog++ (`go-hyperloglog`)

## 1. Mathematical Formulation
Let $p$ be the precision bit length ($4 \le p \le 18$) and $m = 2^p$ be the number of registers.
For each element $x$:
1. Compute 64-bit hash $h = \text{FNV-1a}(x)$.
2. Register index $j = h \gg (64 - p)$.
3. Value $w = h \ll p$. Run of leading zeros $\rho(w) = \text{bits.LeadingZeros64}(w) + 1$.
4. Register update: $M[j] \leftarrow \max(M[j], \rho(w))$.

## 2. Estimator Formula
Raw harmonic mean estimate:
$$E = \alpha_m \cdot m^2 \cdot \left(\sum_{j=0}^{m-1} 2^{-M[j]}\right)^{-1}$$
where:
$$\alpha_m = \begin{cases} 0.673 & m = 16 \\ 0.697 & m = 32 \\ 0.709 & m = 64 \\ \frac{0.7213}{1 + 1.079/m} & m \ge 128 \end{cases}$$

## 3. Range Corrections
- **LinearCounting (Small Range)**: If $E \le 2.5 \cdot m$ and $V > 0$ (count of zero registers):
  $$E^* = m \cdot \ln\left(\frac{m}{V}\right)$$
- **Standard Error**:
  $$\text{SE} \approx \frac{1.04}{\sqrt{m}}$$
  For $p=14$, $m = 16,384 \implies \text{SE} \approx 0.81\%$.
