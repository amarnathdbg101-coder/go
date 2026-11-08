package gobatcher

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBatcher_FlushBySize(t *testing.T) {
	var totalFlushed atomic.Int64
	var flushCount atomic.Int64

	b, err := New(Config[int]{
		MaxBatchSize:  5,
		FlushInterval: 1 * time.Second, // Long interval, should trigger on size
		FlushFn: func(batch []int) error {
			flushCount.Add(1)
			totalFlushed.Add(int64(len(batch)))
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer b.Stop()

	for i := 0; i < 10; i++ {
		_ = b.Push(i)
	}

	// Give slight delay for worker to process size flushes
	time.Sleep(50 * time.Millisecond)

	if totalFlushed.Load() != 10 {
		t.Errorf("expected 10 flushed items, got %d", totalFlushed.Load())
	}
	if flushCount.Load() != 2 {
		t.Errorf("expected 2 batch flushes of size 5, got %d", flushCount.Load())
	}
}

func TestBatcher_FlushByInterval(t *testing.T) {
	var totalFlushed atomic.Int64

	b, err := New(Config[string]{
		MaxBatchSize:  100, // Large size, should trigger by timer
		FlushInterval: 40 * time.Millisecond,
		FlushFn: func(batch []string) error {
			totalFlushed.Add(int64(len(batch)))
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer b.Stop()

	_ = b.Push("msg1")
	_ = b.Push("msg2")

	// Wait for interval tick
	time.Sleep(100 * time.Millisecond)

	if totalFlushed.Load() != 2 {
		t.Errorf("expected 2 items flushed on timer, got %d", totalFlushed.Load())
	}
}

func TestBatcher_ManualFlush(t *testing.T) {
	var totalFlushed atomic.Int64

	b, err := New(Config[int]{
		MaxBatchSize:  100,
		FlushInterval: 5 * time.Second,
		FlushFn: func(batch []int) error {
			totalFlushed.Add(int64(len(batch)))
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer b.Stop()

	_ = b.Push(1)
	_ = b.Push(2)
	_ = b.Push(3)

	b.Flush() // Synchronous manual flush

	if totalFlushed.Load() != 3 {
		t.Errorf("expected 3 items flushed manually, got %d", totalFlushed.Load())
	}
}

func TestBatcher_TryPushBackpressure(t *testing.T) {
	blockChan := make(chan struct{})
	flushingStarted := make(chan struct{})

	b, err := New(Config[int]{
		MaxBatchSize:  1,
		FlushInterval: 10 * time.Second,
		QueueCapacity: 1,
		FlushFn: func(batch []int) error {
			select {
			case flushingStarted <- struct{}{}:
			default:
			}
			<-blockChan
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// First item triggers flush which blocks on blockChan
	_ = b.Push(1)
	<-flushingStarted

	// Second item fills queue (QueueCapacity = 1)
	err = b.TryPush(2)
	if err != nil {
		t.Fatalf("expected item 2 to fill queue, got %v", err)
	}

	// Third item must fail with ErrQueueFull
	err = b.TryPush(3)
	if err != ErrQueueFull {
		t.Errorf("expected ErrQueueFull on saturated buffer, got %v", err)
	}

	close(blockChan)
	b.Stop()
}

func TestBatcher_GracefulStopNoLoss(t *testing.T) {
	var totalFlushed atomic.Int64

	b, err := New(Config[int]{
		MaxBatchSize:  50,
		FlushInterval: 10 * time.Second,
		FlushFn: func(batch []int) error {
			totalFlushed.Add(int64(len(batch)))
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	for i := 0; i < 27; i++ {
		_ = b.Push(i)
	}

	b.Stop() // Must flush remaining 27 items on shutdown

	if totalFlushed.Load() != 27 {
		t.Errorf("expected 27 items flushed on shutdown, got %d", totalFlushed.Load())
	}
}

func TestBatcher_PanicIsolation(t *testing.T) {
	var panicCaught atomic.Bool
	var wg sync.WaitGroup
	wg.Add(1)

	b, err := New(Config[int]{
		MaxBatchSize:  1,
		FlushInterval: 10 * time.Millisecond,
		OnPanic: func(r any) {
			panicCaught.Store(true)
			wg.Done()
		},
		FlushFn: func(batch []int) error {
			panic("simulated flush panic")
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer b.Stop()

	_ = b.Push(99)
	wg.Wait()

	if !panicCaught.Load() {
		t.Errorf("expected flush panic to be caught")
	}
}
