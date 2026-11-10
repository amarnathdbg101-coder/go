package trafficshaper_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-trafficshaper"
)

func TestShaper_BurstAdmission(t *testing.T) {
	shaper, err := trafficshaper.New(
		trafficshaper.WithRate(10.0),
		trafficshaper.WithBurst(5),
		trafficshaper.WithQueueCapacity(10),
	)
	if err != nil {
		t.Fatalf("failed to create shaper: %v", err)
	}
	defer shaper.Close()

	ctx := context.Background()

	// Initial burst of 5 should acquire immediately
	start := time.Now()
	for i := 0; i < 5; i++ {
		if err := shaper.Acquire(ctx); err != nil {
			t.Fatalf("burst acquire failed: %v", err)
		}
	}
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Fatalf("burst took too long: %v", elapsed)
	}

	stats := shaper.Stats()
	if stats.Admitted != 5 {
		t.Fatalf("expected 5 admitted, got %d", stats.Admitted)
	}
}

func TestShaper_QueueFullDrop(t *testing.T) {
	shaper, err := trafficshaper.New(
		trafficshaper.WithRate(1.0), // Very slow: 1 req/s
		trafficshaper.WithBurst(1),
		trafficshaper.WithQueueCapacity(2),
	)
	if err != nil {
		t.Fatalf("failed to create shaper: %v", err)
	}
	defer shaper.Close()

	ctx := context.Background()

	// 1st consumes burst
	if err := shaper.Acquire(ctx); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	// Next 2 fill the queue
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = shaper.Acquire(ctx)
		}()
	}

	// Give goroutines time to enqueue
	time.Sleep(50 * time.Millisecond)

	// 4th request must be rejected with ErrQueueFull
	err = shaper.Acquire(ctx)
	if err != trafficshaper.ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}

	stats := shaper.Stats()
	if stats.Dropped < 1 {
		t.Fatalf("expected dropped count >= 1, got %d", stats.Dropped)
	}
}

func TestShaper_ContextCancellation(t *testing.T) {
	shaper, err := trafficshaper.New(
		trafficshaper.WithRate(1.0),
		trafficshaper.WithBurst(1),
		trafficshaper.WithQueueCapacity(10),
	)
	if err != nil {
		t.Fatalf("failed to create shaper: %v", err)
	}
	defer shaper.Close()

	ctx := context.Background()
	_ = shaper.Acquire(ctx) // consume token

	// Acquire with short context timeout
	shortCtx, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()

	err = shaper.Acquire(shortCtx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestShaper_PriorityScheduling(t *testing.T) {
	shaper, err := trafficshaper.New(
		trafficshaper.WithRate(10.0), // 1 token every 100ms
		trafficshaper.WithBurst(1),
		trafficshaper.WithQueueCapacity(10),
	)
	if err != nil {
		t.Fatalf("failed to create shaper: %v", err)
	}
	defer shaper.Close()

	ctx := context.Background()
	_ = shaper.Acquire(ctx) // drain burst token

	order := make([]int, 0, 2)
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(2)

	// Low priority (0) launched first
	go func() {
		defer wg.Done()
		_ = shaper.AcquireWithPriority(ctx, 0)
		mu.Lock()
		order = append(order, 0)
		mu.Unlock()
	}()

	time.Sleep(20 * time.Millisecond)

	// High priority (10) launched second, should jump ahead
	go func() {
		defer wg.Done()
		_ = shaper.AcquireWithPriority(ctx, 10)
		mu.Lock()
		order = append(order, 10)
		mu.Unlock()
	}()

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 || order[0] != 10 || order[1] != 0 {
		t.Fatalf("priority ordering failed, got: %v", order)
	}
}

