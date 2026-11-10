# Specification: Traffic Shaper Engine (`go-trafficshaper`)

## 1. Problem Statement
Standard rate-limiters immediately reject requests exceeding an instantaneous threshold (returning HTTP 429). In asynchronous message dispatching, webhook deliveries, and API integrations, abrupt drops degrade user experience. A **traffic shaper** buffers bursts in a bounded priority queue and drains them at an exact configured rate.

## 2. Invariants & Mechanics
- **Refill Formula**:
  $$\text{tokens}(t) = \min\left(\text{Burst}, \text{tokens}(t_0) + (t - t_0) \times \text{Rate}\right)$$
- **Admission**:
  - If $\text{tokens} \ge 1.0$ and queue is empty, consume $1.0$ token immediately.
  - If $\text{tokens} < 1.0$, caller enters priority min-heap:
    $$\text{Ticket} = (\text{Priority}, \text{SequenceID})$$
    Tickets with higher priority are dispatched first; ties are resolved by FIFO $\text{SequenceID}$.
- **Bounded Buffer Overflow**:
  - When $\text{len}(\text{Queue}) \ge \text{QueueCapacity}$, additional incoming acquisitions immediately fail with `ErrQueueFull`.
- **Context Cancellation**:
  - A queued caller with an expiring context or explicit cancellation marks its ticket `canceled = true` and exits without consuming future tokens.
