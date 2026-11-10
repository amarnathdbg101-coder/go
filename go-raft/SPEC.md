# Specification: Raft Consensus Protocol (`go-raft`)

## 1. Safety Invariants
1. **Election Safety**: At most one leader can be elected in a given term.
2. **Leader Append-Only**: A leader never overwrites or truncates its entries; it only appends new entries.
3. **Log Matching Property**: If two logs contain an entry with the same index and term, then the logs are identical in all entries up through the given index.
4. **Leader Completeness**: If a log entry is committed in a given term, that entry will be present in the logs of the leaders for all higher-numbered terms.
5. **State Machine Safety**: If a server has applied a log entry at a given index to its state machine, no other server will ever apply a different log entry for the same index.

## 2. Quorum Rules
For a cluster of $N$ nodes:
$$\text{Quorum} = \left\lfloor \frac{N}{2} \right\rfloor + 1$$
- Leader election requires $\ge \text{Quorum}$ votes.
- Commit index advances to $I$ when $\ge \text{Quorum}$ nodes have $\text{matchIndex}[p] \ge I$ and $\text{log}[I].\text{Term} == \text{currentTerm}$.
