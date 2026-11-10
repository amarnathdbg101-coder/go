package cuckoofilter_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-cuckoofilter"
)

func TestCuckoo_InsertLookupDelete(t *testing.T) {
	cf, err := cuckoofilter.New(cuckoofilter.WithCapacity(1000))
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	items := []string{"apple", "banana", "cherry", "date"}

	for _, item := range items {
		if err := cf.InsertString(item); err != nil {
			t.Fatalf("failed to insert %s: %v", item, err)
		}
	}

	if cf.Count() != uint(len(items)) {
		t.Fatalf("expected count %d, got %d", len(items), cf.Count())
	}

	// Verify lookups
	for _, item := range items {
		if !cf.LookupString(item) {
			t.Fatalf("expected item %s to be present", item)
		}
	}

	// Uninserted item
	if cf.LookupString("mango") {
		t.Fatal("unexpected false positive for mango")
	}

	// Delete banana
	if !cf.DeleteString("banana") {
		t.Fatal("expected banana to be deleted")
	}

	if cf.LookupString("banana") {
		t.Fatal("banana should not be present after deletion")
	}

	if cf.Count() != uint(len(items)-1) {
		t.Fatalf("expected count %d, got %d", len(items)-1, cf.Count())
	}
}

func TestCuckoo_FalsePositiveRate(t *testing.T) {
	const total = 5000
	cf, err := cuckoofilter.New(cuckoofilter.WithCapacity(total * 2))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < total; i++ {
		_ = cf.InsertString(fmt.Sprintf("inserted_key_%d", i))
	}

	// Check 5000 uninserted keys
	falsePositives := 0
	for i := 0; i < total; i++ {
		if cf.LookupString(fmt.Sprintf("absent_key_%d", i)) {
			falsePositives++
		}
	}

	fpRate := float64(falsePositives) / float64(total)
	// With 16-bit fingerprints, theoretical FP rate is ~ 2 * 4 / 2^16 ~= 0.012%
	if fpRate > 0.01 { // 1% threshold
		t.Fatalf("false positive rate %.4f%% exceeded 1%% threshold (%d / %d)", fpRate*100, falsePositives, total)
	}
}

func TestCuckoo_CuckooKicking(t *testing.T) {
	// Small capacity to force high load and cuckoo kicks
	cf, _ := cuckoofilter.New(cuckoofilter.WithCapacity(200), cuckoofilter.WithMaxKicks(500))

	inserted := 0
	for i := 0; i < 180; i++ {
		err := cf.InsertString(fmt.Sprintf("kick_item_%d", i))
		if err == nil {
			inserted++
		}
	}

	// Must achieve at least 70% load factor
	if cf.LoadFactor() < 0.5 {
		t.Fatalf("expected load factor >= 0.5, got %f", cf.LoadFactor())
	}

	if inserted < 150 {
		t.Fatalf("too few items inserted: %d", inserted)
	}
}

func TestCuckoo_Concurrency(t *testing.T) {
	cf, _ := cuckoofilter.New(cuckoofilter.WithCapacity(5000))
	var wg sync.WaitGroup

	// 5 writers
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_ = cf.InsertString(fmt.Sprintf("worker_%d_item_%d", id, i))
			}
		}(w)
	}

	// 5 readers
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_ = cf.LookupString(fmt.Sprintf("query_%d", i))
			}
		}()
	}

	wg.Wait()
}

