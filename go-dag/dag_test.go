package dag

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestDAG_LinearExecution(t *testing.T) {
	g := New()
	var sequence []string

	_ = g.AddNode("A", func(ctx context.Context) error {
		sequence = append(sequence, "A")
		return nil
	})
	_ = g.AddNode("B", func(ctx context.Context) error {
		sequence = append(sequence, "B")
		return nil
	})
	_ = g.AddNode("C", func(ctx context.Context) error {
		sequence = append(sequence, "C")
		return nil
	})

	_ = g.AddEdge("A", "B")
	_ = g.AddEdge("B", "C")

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected sort error: %v", err)
	}
	if len(order) != 3 || order[0] != "A" || order[1] != "B" || order[2] != "C" {
		t.Fatalf("unexpected topological order: %v", order)
	}

	err = g.Run(context.Background(), 2)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if len(sequence) != 3 || sequence[0] != "A" || sequence[1] != "B" || sequence[2] != "C" {
		t.Fatalf("unexpected execution sequence: %v", sequence)
	}
}

func TestDAG_DiamondPattern(t *testing.T) {
	g := New()
	var executedCount int32

	_ = g.AddNode("init", func(ctx context.Context) error {
		atomic.AddInt32(&executedCount, 1)
		return nil
	})
	_ = g.AddNode("branch1", func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&executedCount, 1)
		return nil
	})
	_ = g.AddNode("branch2", func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&executedCount, 1)
		return nil
	})
	_ = g.AddNode("final", func(ctx context.Context) error {
		// Both branches must have finished before final
		if atomic.LoadInt32(&executedCount) != 3 {
			return errors.New("final started prematurely")
		}
		atomic.AddInt32(&executedCount, 1)
		return nil
	})

	_ = g.AddEdge("init", "branch1")
	_ = g.AddEdge("init", "branch2")
	_ = g.AddEdge("branch1", "final")
	_ = g.AddEdge("branch2", "final")

	err := g.Run(context.Background(), 4)
	if err != nil {
		t.Fatalf("diamond run failed: %v", err)
	}

	if atomic.LoadInt32(&executedCount) != 4 {
		t.Fatalf("expected all 4 nodes to run, got %d", executedCount)
	}
}

func TestDAG_CycleDetection(t *testing.T) {
	g := New()
	dummy := func(ctx context.Context) error { return nil }

	_ = g.AddNode("1", dummy)
	_ = g.AddNode("2", dummy)
	_ = g.AddNode("3", dummy)

	_ = g.AddEdge("1", "2")
	_ = g.AddEdge("2", "3")
	_ = g.AddEdge("3", "1") // cycle!

	_, err := g.TopologicalSort()
	if !errors.Is(err, ErrCycleDetected) {
		t.Fatalf("expected ErrCycleDetected, got: %v", err)
	}

	err = g.Run(context.Background(), 2)
	if !errors.Is(err, ErrCycleDetected) {
		t.Fatalf("expected ErrCycleDetected on Run, got: %v", err)
	}
}

func TestDAG_TaskFailureAndCancellation(t *testing.T) {
	g := New()
	customErr := errors.New("upstream failed")

	_ = g.AddNode("root", func(ctx context.Context) error {
		return customErr
	})
	_ = g.AddNode("child", func(ctx context.Context) error {
		t.Errorf("child node should not execute after root failed")
		return nil
	})
	_ = g.AddEdge("root", "child")

	err := g.Run(context.Background(), 2)
	if err == nil || !errors.Is(err, customErr) {
		t.Fatalf("expected upstream error, got: %v", err)
	}
}

func TestDAG_PanicRecovery(t *testing.T) {
	g := New()

	_ = g.AddNode("panic_node", func(ctx context.Context) error {
		panic("fatal boom!")
	})

	err := g.Run(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error from node panic")
	}
}
