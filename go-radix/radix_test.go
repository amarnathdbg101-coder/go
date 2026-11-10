package radix_test

import (
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-radix"
)

func TestRadix_EdgeSplittingAndGet(t *testing.T) {
	tr := radix.New[int]()

	words := map[string]int{
		"romane":      1,
		"romanus":     2,
		"romulus":     3,
		"rubens":      4,
		"ruber":       5,
		"rubicon":     6,
		"rubicundus":  7,
	}

	for k, v := range words {
		tr.Insert(k, v)
	}

	if tr.Len() != len(words) {
		t.Fatalf("expected length %d, got %d", len(words), tr.Len())
	}

	// Verify all keys can be fetched
	for k, expectedVal := range words {
		val, ok := tr.Get(k)
		if !ok || val != expectedVal {
			t.Fatalf("key %s: expected %d, got %d (found=%v)", k, expectedVal, val, ok)
		}
	}

	// Non-existent key
	_, ok := tr.Get("roman")
	if ok {
		t.Fatal("prefix 'roman' should not have a value")
	}
}

func TestRadix_LongestPrefix(t *testing.T) {
	tr := radix.New[string]()

	tr.Insert("/api", "API Root")
	tr.Insert("/api/v1", "API v1")
	tr.Insert("/api/v1/orders", "Orders Endpoint")

	// Match exact
	prefix, val, ok := tr.LongestPrefix("/api/v1/orders")
	if !ok || prefix != "/api/v1/orders" || val != "Orders Endpoint" {
		t.Fatalf("expected exact match on orders, got %s -> %s", prefix, val)
	}

	// Match longer path to closest prefix
	prefix, val, ok = tr.LongestPrefix("/api/v1/orders/12345/items")
	if !ok || prefix != "/api/v1/orders" || val != "Orders Endpoint" {
		t.Fatalf("expected longest prefix /api/v1/orders, got %s -> %s", prefix, val)
	}

	// Match shorter
	prefix, val, ok = tr.LongestPrefix("/api/v1/users")
	if !ok || prefix != "/api/v1" || val != "API v1" {
		t.Fatalf("expected prefix /api/v1, got %s -> %s", prefix, val)
	}
}

func TestRadix_WalkPrefix(t *testing.T) {
	tr := radix.New[int]()

	tr.Insert("apple", 1)
	tr.Insert("app", 2)
	tr.Insert("application", 3)
	tr.Insert("aptitude", 4)
	tr.Insert("banana", 5)

	var matches []string
	tr.WalkPrefix("app", func(key string, val int) bool {
		matches = append(matches, key)
		return true
	})

	if len(matches) != 3 {
		t.Fatalf("expected 3 matches for 'app', got %d: %v", len(matches), matches)
	}
	expected := []string{"app", "apple", "application"}
	for i := range expected {
		if matches[i] != expected[i] {
			t.Fatalf("match %d mismatch: expected %s, got %s", i, expected[i], matches[i])
		}
	}
}

func TestRadix_Delete(t *testing.T) {
	tr := radix.New[int]()

	tr.Insert("cat", 10)
	tr.Insert("caterpillar", 20)

	val, deleted := tr.Delete("cat")
	if !deleted || val != 10 {
		t.Fatalf("failed to delete 'cat': val=%d, deleted=%v", val, deleted)
	}

	if tr.Len() != 1 {
		t.Fatalf("expected length 1, got %d", tr.Len())
	}

	// Caterpillar should still be present
	v, ok := tr.Get("caterpillar")
	if !ok || v != 20 {
		t.Fatalf("caterpillar should remain intact, got: %d (ok=%v)", v, ok)
	}
}

func TestRadix_Concurrency(t *testing.T) {
	tr := radix.New[int]()
	var wg sync.WaitGroup

	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				tr.Insert("route/test/path", i)
				_, _ = tr.Get("route/test/path")
				_, _, _ = tr.LongestPrefix("route/test/path/sub")
			}
		}(w)
	}

	wg.Wait()
}

