package dataloader_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-dataloader"
)

func TestBatchCoalescing(t *testing.T) {
	var batchCalls int32

	batchFn := func(ctx context.Context, keys []string) ([]dataloader.Result[string], error) {
		atomic.AddInt32(&batchCalls, 1)
		results := make([]dataloader.Result[string], len(keys))
		for i, k := range keys {
			results[i] = dataloader.Result[string]{Data: "val:" + k}
		}
		return results, nil
	}

	loader := dataloader.New(batchFn, dataloader.WithWait(5*time.Millisecond))
	defer loader.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	keys := []string{"k1", "k2", "k3", "k4", "k5"}
	results := make([]string, len(keys))

	// Launch concurrent requests
	for i, k := range keys {
		wg.Add(1)
		go func(idx int, key string) {
			defer wg.Done()
			val, err := loader.Load(ctx, key)
			if err != nil {
				t.Errorf("unexpected error for %s: %v", key, err)
			}
			results[idx] = val
		}(i, k)
	}

	wg.Wait()

	calls := atomic.LoadInt32(&batchCalls)
	if calls != 1 {
		t.Fatalf("expected exactly 1 batch call for coalesced keys, got %d", calls)
	}

	for i, k := range keys {
		expected := "val:" + k
		if results[i] != expected {
			t.Errorf("key %s: expected %s, got %s", k, expected, results[i])
		}
	}
}

func TestMemoizationCache(t *testing.T) {
	var batchCalls int32

	batchFn := func(ctx context.Context, keys []int) ([]dataloader.Result[int], error) {
		atomic.AddInt32(&batchCalls, 1)
		res := make([]dataloader.Result[int], len(keys))
		for i, k := range keys {
			res[i] = dataloader.Result[int]{Data: k * 10}
		}
		return res, nil
	}

	loader := dataloader.New(batchFn, dataloader.WithWait(2*time.Millisecond))
	defer loader.Close()

	ctx := context.Background()

	// 1st load
	val1, err := loader.Load(ctx, 42)
	if err != nil || val1 != 420 {
		t.Fatalf("first load failed: val=%d err=%v", val1, err)
	}

	// 2nd load (should hit cache)
	val2, err := loader.Load(ctx, 42)
	if err != nil || val2 != 420 {
		t.Fatalf("cached load failed: val=%d err=%v", val2, err)
	}

	calls := atomic.LoadInt32(&batchCalls)
	if calls != 1 {
		t.Fatalf("expected 1 batch call due to caching, got %d", calls)
	}

	// Clear key and load again (should trigger another batch call)
	loader.Clear(42)
	val3, err := loader.Load(ctx, 42)
	if err != nil || val3 != 420 {
		t.Fatalf("reloaded failed after clear: val=%d err=%v", val3, err)
	}

	callsAfterClear := atomic.LoadInt32(&batchCalls)
	if callsAfterClear != 2 {
		t.Fatalf("expected 2 batch calls after cache eviction, got %d", callsAfterClear)
	}
}

func TestPrimeAndClearAll(t *testing.T) {
	var batchCalls int32

	batchFn := func(ctx context.Context, keys []string) ([]dataloader.Result[string], error) {
		atomic.AddInt32(&batchCalls, 1)
		res := make([]dataloader.Result[string], len(keys))
		for i, k := range keys {
			res[i] = dataloader.Result[string]{Data: k}
		}
		return res, nil
	}

	loader := dataloader.New(batchFn)
	defer loader.Close()

	// Pre-prime cache
	loader.Prime("user:101", "Amardeep")

	val, err := loader.Load(context.Background(), "user:101")
	if err != nil || val != "Amardeep" {
		t.Fatalf("expected primed value 'Amardeep', got %s err=%v", val, err)
	}

	if calls := atomic.LoadInt32(&batchCalls); calls != 0 {
		t.Fatalf("expected 0 batch calls for primed key, got %d", calls)
	}

	loader.ClearAll()
	val2, err := loader.Load(context.Background(), "user:101")
	if err != nil || val2 != "user:101" {
		t.Fatalf("expected fetched value after ClearAll, got %s err=%v", val2, err)
	}

	if calls := atomic.LoadInt32(&batchCalls); calls != 1 {
		t.Fatalf("expected 1 batch call after ClearAll, got %d", calls)
	}
}

func TestLoadMany(t *testing.T) {
	batchFn := func(ctx context.Context, keys []int) ([]dataloader.Result[string], error) {
		res := make([]dataloader.Result[string], len(keys))
		for i, k := range keys {
			res[i] = dataloader.Result[string]{Data: fmt.Sprintf("ID#%d", k)}
		}
		return res, nil
	}

	loader := dataloader.New(batchFn)
	defer loader.Close()

	vals, errs := loader.LoadMany(context.Background(), []int{1, 2, 3})
	if len(vals) != 3 {
		t.Fatalf("expected 3 results, got %d", len(vals))
	}
	for i, v := range vals {
		if errs[i] != nil {
			t.Errorf("unexpected error at %d: %v", i, errs[i])
		}
		expected := fmt.Sprintf("ID#%d", i+1)
		if v != expected {
			t.Errorf("expected %s, got %s", expected, v)
		}
	}
}

func TestMaxBatchImmediateDispatch(t *testing.T) {
	var batchCalls int32
	batchFn := func(ctx context.Context, keys []int) ([]dataloader.Result[int], error) {
		atomic.AddInt32(&batchCalls, 1)
		res := make([]dataloader.Result[int], len(keys))
		for i, k := range keys {
			res[i] = dataloader.Result[int]{Data: k}
		}
		return res, nil
	}

	// Long wait (500ms), but max batch size is 2
	loader := dataloader.New(batchFn, dataloader.WithWait(500*time.Millisecond), dataloader.WithMaxBatch(2))
	defer loader.Close()

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		loader.Load(context.Background(), 10)
	}()
	go func() {
		defer wg.Done()
		loader.Load(context.Background(), 20)
	}()

	wg.Wait()
	duration := time.Since(start)

	// Since max batch was reached, it should complete in under 50ms, way before 500ms
	if duration > 100*time.Millisecond {
		t.Errorf("expected immediate dispatch on MaxBatch hit, took %v", duration)
	}

	if calls := atomic.LoadInt32(&batchCalls); calls != 1 {
		t.Errorf("expected 1 batch call, got %d", calls)
	}
}

func TestContextCancellation(t *testing.T) {
	batchFn := func(ctx context.Context, keys []string) ([]dataloader.Result[string], error) {
		time.Sleep(50 * time.Millisecond)
		res := make([]dataloader.Result[string], len(keys))
		return res, nil
	}

	loader := dataloader.New(batchFn, dataloader.WithWait(10*time.Millisecond))
	defer loader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	_, err := loader.Load(ctx, "timeout_key")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded error, got %v", err)
	}
}

