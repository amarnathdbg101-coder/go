# SPEC: go-workerpool/adaptivepool

## 1. Metadata
- **Module**: `go-workerpool`
- **Package**: `adaptivepool`
- **Location**: `C:\Library\go-workerpool\adaptivepool`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `sync`, `sync/atomic`, `time`, `errors`).

---

## 2. Core Capabilities
1. **Dynamic Elastic Goroutine Scaling**:
   - Maintains baseline core workers (`MinWorkers`) and dynamically scales up to `MaxWorkers` during high burst task pressure.
2. **Automated Idle Worker Eviction**:
   - Surplus workers that exceed baseline `MinWorkers` terminate and release memory after inactivity exceeding `IdleTimeout`.
3. **Task Queue Backpressure & Non-Blocking Submissions**:
   - `Submit(task)` blocks when queue is saturated.
   - `TrySubmit(task)` returns immediately with boolean indicator.
   - `SubmitWithTimeout(task, duration)` bounds queue waiting times.
4. **Leak-Proof Shutdown & Drain**:
   - `Close()` initiates non-blocking queue draining and waits for all workers to finish cleanly.

---

## 3. Public API
```go
type Task func()

type Config struct {
    MinWorkers  int
    MaxWorkers  int
    IdleTimeout time.Duration
    QueueSize   int
}

type Pool struct
func New(cfg Config) *Pool
func (p *Pool) Submit(task Task) error
func (p *Pool) TrySubmit(task Task) bool
func (p *Pool) SubmitWithTimeout(task Task, timeout time.Duration) error
func (p *Pool) ActiveWorkers() int
func (p *Pool) PendingTasks() int
func (p *Pool) Close() error
```
