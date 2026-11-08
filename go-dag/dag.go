package dag

import (
	"context"
	"fmt"
	"sync"
)

// TaskFunc defines the executable workload of a DAG node.
type TaskFunc func(ctx context.Context) error

type node struct {
	id   string
	task TaskFunc
}

// Graph represents a directed acyclic graph of tasks.
type Graph struct {
	mu       sync.RWMutex
	nodes    map[string]*node
	outEdges map[string][]string // from -> []to (nodes that depend on from)
	inEdges  map[string][]string // to -> []from (prerequisites of to)
}

// New creates an empty Graph.
func New() *Graph {
	return &Graph{
		nodes:    make(map[string]*node),
		outEdges: make(map[string][]string),
		inEdges:  make(map[string][]string),
	}
}

// AddNode registers a new task node with a unique ID.
func (g *Graph) AddNode(id string, fn TaskFunc) error {
	if fn == nil {
		return ErrNilTask
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[id]; exists {
		return fmt.Errorf("%w: '%s'", ErrDuplicateNode, id)
	}

	g.nodes[id] = &node{id: id, task: fn}
	return nil
}

// AddEdge declares that the 'from' node must complete before the 'to' node can start.
func (g *Graph) AddEdge(from, to string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.nodes[from]; !ok {
		return fmt.Errorf("%w: from node '%s'", ErrNodeNotFound, from)
	}
	if _, ok := g.nodes[to]; !ok {
		return fmt.Errorf("%w: to node '%s'", ErrNodeNotFound, to)
	}

	g.outEdges[from] = append(g.outEdges[from], to)
	g.inEdges[to] = append(g.inEdges[to], from)
	return nil
}

// TopologicalSort returns a valid sequential execution order using Kahn's algorithm.
// Returns ErrCycleDetected if a circular dependency exists.
func (g *Graph) TopologicalSort() ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Compute in-degrees
	inDegree := make(map[string]int)
	for id := range g.nodes {
		inDegree[id] = len(g.inEdges[id])
	}

	// Queue nodes with 0 in-degree
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var order []string

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)

		for _, neighbor := range g.outEdges[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(g.nodes) {
		return nil, ErrCycleDetected
	}

	return order, nil
}

// Run executes the graph nodes concurrently up to concurrency limit, respecting dependency order.
func (g *Graph) Run(ctx context.Context, concurrency int) error {
	// First verify acyclic
	if _, err := g.TopologicalSort(); err != nil {
		return err
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.nodes) == 0 {
		return nil
	}

	if concurrency <= 0 {
		concurrency = len(g.nodes)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Track in-degrees and completed count
	inDegree := make(map[string]int)
	for id := range g.nodes {
		inDegree[id] = len(g.inEdges[id])
	}

	sem := make(chan struct{}, concurrency)
	readyChan := make(chan string, len(g.nodes))
	errOnce := sync.Once{}
	var runErr error

	var lock sync.Mutex
	var wg sync.WaitGroup

	// Enqueue root nodes
	for id, deg := range inDegree {
		if deg == 0 {
			readyChan <- id
		}
	}

	totalNodes := len(g.nodes)
	processedNodes := 0

	for processedNodes < totalNodes {
		select {
		case <-ctx.Done():
			wg.Wait()
			if runErr != nil {
				return runErr
			}
			return ctx.Err()
		case nodeID := <-readyChan:
			processedNodes++
			wg.Add(1)

			go func(id string) {
				defer wg.Done()

				// Acquire concurrency slot
				select {
				case <-ctx.Done():
					return
				case sem <- struct{}{}:
				}
				defer func() { <-sem }()

				// Execute task with panic recovery
				err := g.executeTask(ctx, id)
				if err != nil {
					errOnce.Do(func() {
						runErr = fmt.Errorf("dag node '%s' failed: %w", id, err)
						cancel() // abort remaining nodes
					})
					return
				}

				// Check dependents
				lock.Lock()
				defer lock.Unlock()

				for _, neighbor := range g.outEdges[id] {
					inDegree[neighbor]--
					if inDegree[neighbor] == 0 {
						readyChan <- neighbor
					}
				}
			}(nodeID)
		}

		if runErr != nil {
			break
		}
	}

	wg.Wait()
	return runErr
}

func (g *Graph) executeTask(ctx context.Context, id string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("task panic: %v", r)
		}
	}()

	taskNode := g.nodes[id]
	return taskNode.task(ctx)
}
