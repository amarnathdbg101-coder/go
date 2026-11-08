package goringbuffer

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestRingBuffer_FIFO(t *testing.T) {
	rb, err := New[int](3, false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_ = rb.Push(10)
	_ = rb.Push(20)
	_ = rb.Push(30)

	if !rb.IsFull() {
		t.Errorf("expected buffer to be full")
	}

	val, err := rb.Peek()
	if err != nil || val != 10 {
		t.Fatalf("expected peek 10, got %d", val)
	}

	// Pop 1
	v, err := rb.Pop()
	if err != nil || v != 10 {
		t.Errorf("expected 10, got %d", v)
	}

	// Pop 2
	v, err = rb.Pop()
	if err != nil || v != 20 {
		t.Errorf("expected 20, got %d", v)
	}

	// Pop 3
	v, err = rb.Pop()
	if err != nil || v != 30 {
		t.Errorf("expected 30, got %d", v)
	}

	if !rb.IsEmpty() {
		t.Errorf("expected buffer to be empty")
	}

	// Pop empty
	_, err = rb.Pop()
	if !errors.Is(err, ErrBufferEmpty) {
		t.Errorf("expected ErrBufferEmpty, got %v", err)
	}
}

func TestRingBuffer_NoOverwriteFullError(t *testing.T) {
	rb, _ := New[string](2, false)

	_ = rb.Push("a")
	_ = rb.Push("b")

	err := rb.Push("c")
	if !errors.Is(err, ErrBufferFull) {
		t.Errorf("expected ErrBufferFull, got %v", err)
	}
}

func TestRingBuffer_Overwrite(t *testing.T) {
	rb, _ := New[string](3, true)

	_ = rb.Push("1")
	_ = rb.Push("2")
	_ = rb.Push("3")
	// Buffer is now [1, 2, 3]

	// Overwrite: pushes 4 (1 should be dropped), then 5 (2 should be dropped)
	_ = rb.Push("4")
	_ = rb.Push("5")

	if rb.Len() != 3 {
		t.Errorf("expected len 3, got %d", rb.Len())
	}

	slice := rb.Slice()
	expected := []string{"3", "4", "5"}
	if !reflect.DeepEqual(slice, expected) {
		t.Errorf("expected %v, got %v", expected, slice)
	}
}

func TestRingBuffer_ConcurrentStress(t *testing.T) {
	rb, _ := New[int](100, true)
	var wg sync.WaitGroup
	workers := 10
	iterations := 500

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = rb.Push(workerID*1000 + j)
				_, _ = rb.Pop()
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkRingBuffer_PushPop(b *testing.B) {
	rb, _ := New[int](1024, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rb.Push(i)
		_, _ = rb.Pop()
	}
}
