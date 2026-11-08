# go-priorityqueue

High-performance, generic binary heap Priority Queue with Max/Min heap modes for Go.

---

## Features
- **O(log n) Operations**: Contiguous slice binary heap with logarithmic push/pop and $O(1)$ peek.
- **MaxHeap & MinHeap**: Configurable ordering — pop urgent tasks first (MaxHeap) or earliest deadlines first (MinHeap).
- **Type-Safe Generics**: Works seamlessly with any payload type `T` without interface assertions.
- **Thread Safe**: Protected with concurrent `sync.RWMutex`.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	gopriorityqueue "go-priorityqueue"
)

type Task struct {
	Name string
}

func main() {
	// Create MaxHeap priority queue
	pq := gopriorityqueue.New[Task]()

	// Push tasks with priority scores
	pq.Push(Task{Name: "Send Newsletter"}, 10)
	pq.Push(Task{Name: "Process Payment"}, 100)
	pq.Push(Task{Name: "Generate PDF Report"}, 50)

	// Pop in priority order (Highest to Lowest)
	for !pq.IsEmpty() {
		task, prio, _ := pq.Pop()
		fmt.Printf("[Priority %d] Executing: %s\n", prio, task.Name)
	}
	// Output:
	// [Priority 100] Executing: Process Payment
	// [Priority 50] Executing: Generate PDF Report
	// [Priority 10] Executing: Send Newsletter
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
