package gopool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type MockConn struct {
	ID      int
	Healthy bool
	Closed  bool
}

func (m *MockConn) Close() error {
	m.Closed = true
	return nil
}

func TestPool_AcquireReleaseReuse(t *testing.T) {
	var createdCount atomic.Int32

	pool, err := New(Config[*MockConn]{
		Capacity: 2,
		Factory: func(ctx context.Context) (*MockConn, error) {
			id := int(createdCount.Add(1))
			return &MockConn{ID: id, Healthy: true}, nil
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer pool.Close()

	// 1. Acquire first connection
	c1, err := pool.Acquire(context.Background())
	if err != nil || c1.ID != 1 {
		t.Fatalf("failed to acquire c1: %v", err)
	}

	// 2. Release back to pool
	pool.Release(c1)

	// 3. Re-acquire should reuse c1 without calling Factory again
	c2, err := pool.Acquire(context.Background())
	if err != nil || c2.ID != 1 {
		t.Fatalf("expected reused connection id 1, got id %d", c2.ID)
	}

	if createdCount.Load() != 1 {
		t.Errorf("expected factory called only once, got %d", createdCount.Load())
	}
	pool.Release(c2)
}

func TestPool_CapacityBoundingAndTimeout(t *testing.T) {
	pool, _ := New(Config[*MockConn]{
		Capacity: 1, // Only 1 resource permitted
		Factory: func(ctx context.Context) (*MockConn, error) {
			return &MockConn{ID: 1, Healthy: true}, nil
		},
	})
	defer pool.Close()

	// Acquire the only connection
	c1, _ := pool.Acquire(context.Background())

	// Attempting 2nd acquire with 30ms timeout should fail with context deadline exceeded
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := pool.Acquire(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	pool.Release(c1)
}

func TestPool_HealthCheck(t *testing.T) {
	var createdCount atomic.Int32

	pool, _ := New(Config[*MockConn]{
		Capacity: 2,
		Factory: func(ctx context.Context) (*MockConn, error) {
			id := int(createdCount.Add(1))
			return &MockConn{ID: id, Healthy: true}, nil
		},
		HealthCheck: func(m *MockConn) bool {
			return m.Healthy
		},
	})
	defer pool.Close()

	c1, _ := pool.Acquire(context.Background())
	// Make connection unhealthy before releasing
	c1.Healthy = false
	pool.Release(c1)

	// Next acquire should detect unhealthy, discard c1, and invoke Factory for c2
	c2, err := pool.Acquire(context.Background())
	if err != nil || c2.ID != 2 {
		t.Fatalf("expected fresh connection id 2, got id %d (err %v)", c2.ID, err)
	}
	if !c1.Closed {
		t.Errorf("unhealthy connection should have been closed")
	}

	pool.Release(c2)
}

func TestPool_IdleTimeout(t *testing.T) {
	var closedCount atomic.Int32

	pool, _ := New(Config[*MockConn]{
		Capacity:    2,
		IdleTimeout: 30 * time.Millisecond,
		Factory: func(ctx context.Context) (*MockConn, error) {
			return &MockConn{ID: 1, Healthy: true}, nil
		},
		CloseFn: func(m *MockConn) error {
			closedCount.Add(1)
			return nil
		},
	})
	defer pool.Close()

	c, _ := pool.Acquire(context.Background())
	pool.Release(c)

	// Sleep longer than IdleTimeout
	time.Sleep(60 * time.Millisecond)

	// Next acquire should evict expired idle connection and create fresh one
	cFresh, _ := pool.Acquire(context.Background())
	if closedCount.Load() != 1 {
		t.Errorf("expected 1 connection closed due to idle timeout, got %d", closedCount.Load())
	}

	pool.Release(cFresh)
}
