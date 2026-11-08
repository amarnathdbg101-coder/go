package cachekit_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-cachekit"
)

type Product struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

// -------------------------------------------------------------------------
// 1. MemoryStore CRUD & TTL Expiry Tests
// -------------------------------------------------------------------------

func TestMemoryStore_BasicCRUD(t *testing.T) {
	ctx := context.Background()
	store := cachekit.NewMemoryStore()
	defer store.Close()

	key := "test:item"
	val := []byte("hello cache")

	// 1. Miss initially
	_, err := store.Get(ctx, key)
	if err != cachekit.ErrCacheMiss {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}

	// 2. Set with TTL
	if err := store.Set(ctx, key, val, 1*time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	// 3. Get should return exact bytes
	got, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if string(got) != string(val) {
		t.Errorf("expected '%s', got '%s'", val, got)
	}

	// 4. Delete
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete error: %v", err)
	}
	_, err = store.Get(ctx, key)
	if err != cachekit.ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss after delete, got %v", err)
	}
}

func TestMemoryStore_Expiration(t *testing.T) {
	ctx := context.Background()
	store := cachekit.NewMemoryStore()
	defer store.Close()

	key := "temp:key"
	_ = store.Set(ctx, key, []byte("temporary"), 50*time.Millisecond)

	// Immediate get succeeds
	if _, err := store.Get(ctx, key); err != nil {
		t.Fatalf("immediate get failed: %v", err)
	}

	// Wait for TTL to expire
	time.Sleep(70 * time.Millisecond)

	// Subsequent get triggers passive eviction
	_, err := store.Get(ctx, key)
	if err != cachekit.ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss after TTL elapsed, got %v", err)
	}
}

func TestMemoryStore_DeletePrefix(t *testing.T) {
	ctx := context.Background()
	store := cachekit.NewMemoryStore()
	defer store.Close()

	_ = store.Set(ctx, "products:1", []byte("p1"), time.Minute)
	_ = store.Set(ctx, "products:2", []byte("p2"), time.Minute)
	_ = store.Set(ctx, "users:1", []byte("u1"), time.Minute)

	if err := store.DeletePrefix(ctx, "products:"); err != nil {
		t.Fatalf("delete prefix failed: %v", err)
	}

	// products should be gone
	if _, err := store.Get(ctx, "products:1"); err != cachekit.ErrCacheMiss {
		t.Errorf("expected products:1 deleted")
	}
	if _, err := store.Get(ctx, "products:2"); err != cachekit.ErrCacheMiss {
		t.Errorf("expected products:2 deleted")
	}

	// users:1 should remain intact
	if _, err := store.Get(ctx, "users:1"); err != nil {
		t.Errorf("expected users:1 to remain")
	}
}

// -------------------------------------------------------------------------
// 2. Cache Stampede / Dogpile Protection (Singleflight Test)
// -------------------------------------------------------------------------

func TestCacheStampede_SingleflightProtection(t *testing.T) {
	ctx := context.Background()
	store := cachekit.NewMemoryStore()
	defer store.Close()

	engine, err := cachekit.New(store)
	if err != nil {
		t.Fatalf("engine new error: %v", err)
	}

	var fetcherCallCount int64
	key := "hot:product:101"

	// 100 concurrent goroutines request the exact same key at the exact same millisecond
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	productCache := cachekit.Typed[Product](engine)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			prod, err := productCache.GetOrSet(ctx, key, 5*time.Minute, func(ctx context.Context) (Product, error) {
				// Simulate slow database query (50ms)
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

	// CRITICAL ASSERTION:
	// Even though 100 goroutines hit at the same time, fetcher must be called EXACTLY ONCE!
	calls := atomic.LoadInt64(&fetcherCallCount)
	if calls != 1 {
		t.Fatalf("Cache Stampede occurred! Expected 1 database fetch, but got %d", calls)
	}
}

// -------------------------------------------------------------------------
// 3. TwoTierStore L1 + L2 Hybrid Caching Tests
// -------------------------------------------------------------------------

func TestTwoTierStore_Flow(t *testing.T) {
	ctx := context.Background()

	l1 := cachekit.NewMemoryStore()
	defer l1.Close()

	l2 := cachekit.NewMemoryStore() // Simulating remote L2 cache
	defer l2.Close()

	twoTier := cachekit.NewTwoTierStore(l1, l2)
	defer twoTier.Close()

	key := "hybrid:key"
	val := []byte("hybrid value")

	// 1. Set writes to both L1 and L2
	if err := twoTier.Set(ctx, key, val, 10*time.Minute); err != nil {
		t.Fatalf("twoTier set error: %v", err)
	}

	// Verify it exists in both
	if _, err := l1.Get(ctx, key); err != nil {
		t.Errorf("expected key in L1")
	}
	if _, err := l2.Get(ctx, key); err != nil {
		t.Errorf("expected key in L2")
	}

	// 2. Clear L1 (simulate cold restart of application node)
	_ = l1.Delete(ctx, key)
	if _, err := l1.Get(ctx, key); err != cachekit.ErrCacheMiss {
		t.Fatalf("expected L1 to be empty")
	}

	// 3. TwoTier Get should fall back to L2 AND auto backfill L1!
	got, err := twoTier.Get(ctx, key)
	if err != nil {
		t.Fatalf("twoTier get failed: %v", err)
	}
	if string(got) != string(val) {
		t.Errorf("expected '%s', got '%s'", val, got)
	}

	// 4. L1 must now be populated from backfill!
	backfilled, err := l1.Get(ctx, key)
	if err != nil || string(backfilled) != string(val) {
		t.Errorf("expected L1 to be backfilled from L2")
	}

	// 5. Delete removes from both
	_ = twoTier.Delete(ctx, key)
	if _, err := l1.Get(ctx, key); err != cachekit.ErrCacheMiss {
		t.Errorf("expected L1 evicted")
	}
	if _, err := l2.Get(ctx, key); err != cachekit.ErrCacheMiss {
		t.Errorf("expected L2 evicted")
	}
}

// -------------------------------------------------------------------------
// 4. Generics Type-Safe Struct Tests
// -------------------------------------------------------------------------

func TestGenerics_TypedCache(t *testing.T) {
	ctx := context.Background()
	store := cachekit.NewMemoryStore()
	defer store.Close()

	engine, _ := cachekit.New(store)

	prod := Product{ID: "p_10", Title: "Mechanical Keyboard", Price: 3499.0}
	productCache := cachekit.Typed[Product](engine)

	// Set & Get
	if err := productCache.Set(ctx, "prod:10", prod, time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	fetched, err := productCache.Get(ctx, "prod:10")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if fetched.ID != prod.ID || fetched.Price != prod.Price {
		t.Errorf("struct mismatch, got %+v", fetched)
	}
}

func TestEngine_NilStoreCheck(t *testing.T) {
	_, err := cachekit.New(nil)
	if err != cachekit.ErrStoreNil {
		t.Errorf("expected ErrStoreNil, got %v", err)
	}
}

func TestGetOrSetSWR_Flow(t *testing.T) {
	ctx := context.Background()
	store := cachekit.NewMemoryStore()
	defer store.Close()

	engine, _ := cachekit.New(store)

	key := "swr:product:201"
	var callCount int64

	fetcher := func(ctx context.Context) (Product, error) {
		atomic.AddInt64(&callCount, 1)
		return Product{ID: "201", Title: "Wireless Mouse", Price: 1299.0}, nil
	}

	// 1. Initial call (synchronous miss) -> fetcher called (count = 1)
	val1, err := cachekit.GetOrSetSWR[Product](ctx, engine, key, 50*time.Millisecond, 500*time.Millisecond, fetcher)
	if err != nil || val1.ID != "201" {
		t.Fatalf("swr initial get failed: %v", err)
	}
	if atomic.LoadInt64(&callCount) != 1 {
		t.Errorf("expected fetcher count 1, got %d", atomic.LoadInt64(&callCount))
	}

	// 2. Immediate second call (fresh) -> fetcher NOT called (count = 1)
	val2, _ := cachekit.GetOrSetSWR[Product](ctx, engine, key, 50*time.Millisecond, 500*time.Millisecond, fetcher)
	if val2.ID != "201" || atomic.LoadInt64(&callCount) != 1 {
		t.Errorf("expected fresh hit without extra fetcher call")
	}

	// 3. Wait 70ms for freshTTL (50ms) to expire, so key becomes STALE
	time.Sleep(70 * time.Millisecond)

	// 4. Stale call -> returns stale value instantly (0ms), triggers async background refresh
	val3, _ := cachekit.GetOrSetSWR[Product](ctx, engine, key, 50*time.Millisecond, 500*time.Millisecond, fetcher)
	if val3.ID != "201" {
		t.Errorf("expected stale value returned immediately")
	}

	// Wait for background async refresh goroutine to finish
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt64(&callCount) != 2 {
		t.Errorf("expected async fetcher count 2 after SWR revalidation, got %d", atomic.LoadInt64(&callCount))
	}
}

func BenchmarkMemoryStore_Get(b *testing.B) {
	ctx := context.Background()
	store := cachekit.NewMemoryStore()
	defer store.Close()

	_ = store.Set(ctx, "bench:key", []byte("benchmark value"), 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get(ctx, "bench:key")
	}
}

func BenchmarkSingleflight_GetOrSet(b *testing.B) {
	ctx := context.Background()
	store := cachekit.NewMemoryStore()
	defer store.Close()

	engine, _ := cachekit.New(store)
	_ = cachekit.Set(ctx, engine, "bench:prod", Product{ID: "1", Title: "Bench", Price: 100}, 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cachekit.GetOrSet(ctx, engine, "bench:prod", 10*time.Minute, func(ctx context.Context) (Product, error) {
			return Product{ID: "1", Title: "Bench", Price: 100}, nil
		})
	}
}
