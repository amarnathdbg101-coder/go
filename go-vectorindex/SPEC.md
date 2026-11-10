# Specification: Vector Similarity Search Index (`go-vectorindex`)

## 1. Mathematical Formulations
For query vector $\mathbf{q} \in \mathbb{R}^D$ and candidate vector $\mathbf{v} \in \mathbb{R}^D$:

1. **Cosine Similarity**:
   $$\text{Cosine}(\mathbf{q}, \mathbf{v}) = \frac{\mathbf{q} \cdot \mathbf{v}}{\|\mathbf{q}\|_2 \|\mathbf{v}\|_2} = \frac{\sum_{i=1}^D q_i v_i}{\sqrt{\sum_{i=1}^D q_i^2} \sqrt{\sum_{i=1}^D v_i^2}}$$
2. **Euclidean Distance ($L_2$)**:
   $$\text{Distance}_{L_2}(\mathbf{q}, \mathbf{v}) = \sqrt{\sum_{i=1}^D (q_i - v_i)^2}$$
3. **Dot Product**:
   $$\mathbf{q} \cdot \mathbf{v} = \sum_{i=1}^D q_i v_i$$

## 2. Invariants & Complexity
- **Vector Space Dimension**: All inserted vectors must strictly have length equal to `Dimensions`.
- **Heap Maintenance**:
  - For similarity metrics, a **Min-Heap** of size $K$ retains the $K$ largest similarities seen so far.
  - For distance metrics, a **Max-Heap** of size $K$ retains the $K$ shortest distances seen so far.
- **Search Complexity**: $O(N \cdot D + N \cdot \log K)$ where $N$ is vector count, $D$ is dimensionality, and $K$ is nearest neighbor count.
