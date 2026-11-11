# SPEC: go-async

## 1. Scope & Purpose
Unified asynchronous processing, worker pools, event channels, and workflow pipelines suite for Go microservices.
Consolidates Bounded Worker Pools, Task Queue with Retries, PubSub Fanout, Event Bus, Functional Parallel Streams, and Cron Scheduler under one clean, modular package.

## 2. Public API Sub-packages
- `github.com/shopsilo/go-async/taskqueue`: Bounded worker pool & panic isolation task queue
- `github.com/shopsilo/go-async/workerpool`: Bounded goroutine pool with backpressure
- `github.com/shopsilo/go-async/events`: In-memory event bus with wildcards & DLQ
- `github.com/shopsilo/go-async/pubsub`: Non-blocking channel fanout

## 3. Concurrency & Memory Guarantees
- Bounded memory buffers & goroutines.
- Panic isolation prevents worker pool crashes.
- Graceful shutdown via `Close()`.
