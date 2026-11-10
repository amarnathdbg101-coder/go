# Specification: Locality Sensitive Hashing (`go-simhash`)

## 1. Theoretical Foundation
Traditional cryptographic hash functions (SHA-256, MD5) exhibit the "avalanche effect": changing 1 input bit alters approximately 50% of the output bits.
In contrast, **SimHash** (Charikar, 2002) is a Locality Sensitive Hash where:
$$\Pr[\text{SimHash}(A)_i = \text{SimHash}(B)_i] = 1 - \frac{\theta(A, B)}{\pi}$$
where $\theta(A, B)$ is the angle between document feature vectors.

## 2. Algorithm Steps
1. Tokenize text into $m$ features $F = \{(t_1, w_1), (t_2, w_2), \dots, (t_m, w_m)\}$.
2. Initialize an accumulator vector $V \in \mathbb{Z}^{64}$ with all zeros.
3. For each feature $(t_i, w_i)$:
   - Compute $h_i = \text{FNV-1a}(t_i) \in \{0, 1\}^{64}$.
   - For bit $b \in [0, 63]$:
     $$V_b \leftarrow V_b + \begin{cases} +w_i & \text{if } (h_i)_b = 1 \\ -w_i & \text{if } (h_i)_b = 0 \end{cases}$$
4. Construct fingerprint bitwise:
   $$\text{Fingerprint}_b = \begin{cases} 1 & \text{if } V_b > 0 \\ 0 & \text{otherwise} \end{cases}$$

## 3. Hamming Distance
$$\text{HammingDistance}(F_1, F_2) = \text{bits.OnesCount64}(F_1 \oplus F_2)$$
- Distance $0 \implies$ Identical feature sets.
- Distance $\le 3 \implies$ Near-duplicate document (typical web crawler threshold).
- Distance $\approx 32 \implies$ Completely unrelated random documents.
