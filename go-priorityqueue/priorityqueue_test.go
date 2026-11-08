package gopriorityqueue

import (
	"errors"
	"sync"
	"testing"
)

func TestPriorityQueue_MaxHeap(t *testing.T) {
	pq := New[string]()

	pq.Push("low", 10)
	pq.Push("critical", 100)
	pq.Push("medium", 50)
	pq.Push("urgent", 90)

	if pq.Len() != 4 {
		t.Fatalf("expected len 4, got %d", pq.Len())
	}

	val, prio, err := pq.Peek()
	if err != nil || val != "critical" || prio != 100 {
		t.Fatalf("expected peek critical(100), got %s(%d)", val, prio)
	}

	// Pop sequence should be 100 -> 90 -> 50 -> 10
	expected := []struct {
		val  string
		prio int
	}{
		{"critical", 100},
		{"urgent", 90},
		{"medium", 50},
		{"low", 10},
	}

	for _, exp := range expected {
		v, p, err := pq.Pop()
		if err != nil || v != exp.val || p != exp.prio {
			t.Errorf("expected %s(%d), got %s(%d), err %v", exp.val, exp.prio, v, p, err)
		}
	}

	// Empty queue pop
	_, _, err = pq.Pop()
	if !errors.Is(err, ErrQueueEmpty) {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestPriorityQueue_MinHeap(t *testing.T) {
	pq := NewWithOrder[int](MinHeap)

	pq.Push(300, 300)
	pq.Push(10, 10)
	pq.Push(50, 50)

	v, p, _ := pq.Pop()
	if v != 10 || p != 10 {
		t.Errorf("expected min 10, got %d", v)
	}

	v, p, _ = pq.Pop()
	if v != 50 || p != 50 {
		t.Errorf("expected min 50, got %d", v)
	}

	v, p, _ = pq.Pop()
	if v != 300 || p != 300 {
		t.Errorf("expected min 300, got %d", v)
	}
}

func TestPriorityQueue_ConcurrentStress(t *testing.T) {
	pq := New[int]()
	var wg sync.WaitGroup
	workers := 10
	iterations := 200

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				pq.Push(w*1000+j, j)
			}
		}(i)
	}

	wg.Wait()

	if pq.Len() != workers*iterations {
		t.Fatalf("expected total %d items, got %d", workers*iterations, pq.Len())
	}

	// Verify all items are popped in non-increasing priority order
	prevPrio := 999999
	for !pq.IsEmpty() {
		_, prio, err := pq.Pop()
		if err != nil {
			t.Fatalf("unexpected pop error: %v", err)
		}
		if prio > prevPrio {
			t.Fatalf("heap invariant violated: %d > %d", prio, prevPrio)
		}
		prevPrio = prio
	}
}

func BenchmarkPriorityQueue_PushPop(b *testing.B) {
	pq := New[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Push(i, i%100)
	}
	for i := 0; i < b.N; i++ {
		_, _, _ = pq.Pop()
	}
}
