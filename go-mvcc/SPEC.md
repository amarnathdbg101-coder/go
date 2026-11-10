# Specification: Multi-Version Concurrency Control (`go-mvcc`)

## 1. Snapshot Visibility Rules
For transaction $T$ with identifier $\text{TxID}(T)$ and active snapshot set $\text{Active}(T)$:

A version $V$ created by $\text{TxID}(V)$ and deleted by $\text{DeletedTxID}(V)$ is visible to $T$ if and only if:
1. $V$ is committed ($\text{Committed}(V) = \text{true}$).
2. $\text{TxID}(V) \le \text{TxID}(T)$ and $\text{TxID}(V) \notin \text{Active}(T)$.
3. Either:
   - $V$ is not deleted ($\text{DeletedTxID}(V) = 0$), OR
   - The deleting transaction was after $T$'s snapshot ($\text{DeletedTxID}(V) > \text{TxID}(T)$ or $\text{DeletedTxID}(V) \in \text{Active}(T)$).

## 2. Conflict Invariant (First-Committer-Wins)
When transaction $T$ commits a write to key $K$:
$$\text{Conflict} \iff \exists V \in \text{Chain}(K) \text{ such that } \text{TxID}(V) > \text{TxID}(T)$$
If conflict exists, $T$ aborts with `ErrWriteConflict`.
