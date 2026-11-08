# SPEC: go-dag

## 1. Metadata
- **Module**: `go-dag`
- **Package**: `dag`
- **Location**: `C:\Library\go-dag`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `context`, `sync`, `errors`, `fmt`).

---

## 2. Core Capabilities
1. **Cycle Detection & Topological Sorting**:
   - Implements Kahn's algorithm to calculate linear execution dependencies and reliably detect circular deadlock graphs (`ErrCycleDetected`).
2. **Concurrency-Bounded Parallel Execution**:
   - `Run(ctx, concurrency)` schedules ready prerequisite-free tasks to execute concurrently across worker routines up to the configured limit.
3. **Cascading Failure & Panic Isolation**:
   - Recovers from node panics without crashing the host process.
   - Automatically aborts remaining running tasks via context propagation upon node failure.

---

## 3. Public API
```go
type TaskFunc func(ctx context.Context) error

type Graph struct
func New() *Graph
func (g *Graph) AddNode(id string, fn TaskFunc) error
func (g *Graph) AddEdge(from, to string) error
func (g *Graph) TopologicalSort() ([]string, error)
func (g *Graph) Run(ctx context.Context, concurrency int) error
```
