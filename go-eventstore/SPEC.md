# Specification: Event Sourcing Engine (`go-eventstore`)

## 1. Stream & Concurrency Invariants
For any stream $S$:
1. **Monotonic Stream Revision**:
   $$\text{Revision}(E_{i+1}) = \text{Revision}(E_i) + 1$$
2. **Optimistic Concurrency Control Invariant**:
   When appending with $\text{ExpectedRevision} = R$:
   - If $R = \text{NoStream} \implies \text{len}(S) = 0$.
   - If $R = \text{AnyVersion} \implies$ unchecked append.
   - Else, append succeeds iff $\text{len}(S) = R$, otherwise rejects with `ErrWrongExpectedVersion`.
3. **Global Ordering**:
   Every recorded event receives a strictly monotonic `GlobalPosition` across all streams in the store ($1, 2, 3, \dots$).
