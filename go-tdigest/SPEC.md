# Specification: t-Digest Quantile Estimator (`go-tdigest`)

## 1. Scale Function & Cluster Invariants
Let $N$ be total weight and $\delta$ be the compression parameter.
For each centroid at quantile $q \in [0, 1]$:
$$\text{MaxWeight}(q) = \frac{4 N q (1 - q)}{\delta}$$
- Near the median ($q = 0.5$): $\text{MaxWeight} \approx N / \delta$.
- Near the extreme tails ($q \to 0$ or $q \to 1$): $\text{MaxWeight} \to 0$.
This mathematical constraint ensures clusters near $0$ and $1$ are forced to remain small (often holding only individual data points), providing high fidelity for tail percentiles.

## 2. Quantile Interpolation
Between two adjacent centroids $C_i = (\mu_i, w_i)$ and $C_{i+1} = (\mu_{i+1}, w_{i+1})$:
$$v = \mu_i + \frac{W_{\text{target}} - W_i}{\frac{w_i}{2} + \frac{w_{i+1}}{2}} (\mu_{i+1} - \mu_i)$$
