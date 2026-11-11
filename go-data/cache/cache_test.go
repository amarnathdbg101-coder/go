package cache_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shopsilo/go-data/cache"
)

type Product struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

func TestMemoryStore_BasicCRUD(t *testing.T) {
	ctx := context.Background()
	store := cache.NewMemoryStore()
	defer store.Close()

	key := "test:item"
	val := []byte("hello cache")

	_, err := store.Get(ctx, key)
	if err != cache.ErrCacheMiss {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}

	if err := store.Set(ctx, key, val, 1*time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	got, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if string(got) != string(val) {
		t.Errorf("expected '%s', got '%s'", val, got)
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete error: %v", err)
	}
	_, err = store.Get(ctx, key)
	if err != cache.ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss after delete, got %v", err)
	}
}

func TestCacheStampede_SingleflightProtection(t *testing.T) {
	ctx := context.Background()
	store := cache.NewMemoryStore()
	defer store.Close()

	engine, err := cache.New(store)
	if err != nil {
		t.Fatalf("engine new error: %v", err)
	}

	var fetcherCallCount int64
	key := "hot:product:101"

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	productCache := cache.Typed[Product](engine)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			prod, err := productCache.GetOrSet(ctx, key, 5*time.Minute, func(ctx context.Context) (Product, error) {
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt64(&fetcherCallCount, 1)
				return Product{ID: "101", Title: "MacBook Pro", Price: 199999.0}, nil
			})
			if err != nil {
				t.Errorf("concurrent getOrSet error: %v", err)
			}
			if prod.ID != "101" {
				t.Errorf("expected product ID 101, got %s", prod.ID)
			}
		}()
	}

	wg.Wait()

	calls := atomic.LoadInt64(&fetcherCallCount)
	if calls != 1 {
		t.Fatalf("Cache Stampede occurred! Expected 1 database fetch, but got %d", calls)
	}
}
