package orderedmap

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

func TestOrderedMap_Basic(t *testing.T) {
	om := New[string, int]()

	om.Set("first", 100)
	om.Set("second", 200)
	om.Set("third", 300)

	if om.Len() != 3 {
		t.Fatalf("expected len 3, got %d", om.Len())
	}

	val, ok := om.Get("second")
	if !ok || val != 200 {
		t.Fatalf("expected 200, got %d, ok: %v", val, ok)
	}

	if !om.Has("third") {
		t.Fatalf("expected third to exist")
	}
	if om.Has("missing") {
		t.Fatalf("did not expect missing to exist")
	}

	// Update existing key - should not change insertion order
	om.Set("first", 999)
	if val, _ := om.Get("first"); val != 999 {
		t.Fatalf("expected 999, got %d", val)
	}

	keys := om.Keys()
	expectedKeys := []string{"first", "second", "third"}
	for i, k := range keys {
		if k != expectedKeys[i] {
			t.Fatalf("index %d: expected key %s, got %s", i, expectedKeys[i], k)
		}
	}

	values := om.Values()
	expectedValues := []int{999, 200, 300}
	for i, v := range values {
		if v != expectedValues[i] {
			t.Fatalf("index %d: expected value %d, got %d", i, expectedValues[i], v)
		}
	}
}

func TestOrderedMap_Delete(t *testing.T) {
	om := New[string, string]()
	om.Set("a", "alpha")
	om.Set("b", "beta")
	om.Set("c", "gamma")
	om.Set("d", "delta")

	// Delete middle
	if !om.Delete("b") {
		t.Fatalf("expected delete b to return true")
	}
	if om.Delete("b") {
		t.Fatalf("expected second delete b to return false")
	}
	if om.Len() != 3 {
		t.Fatalf("expected len 3, got %d", om.Len())
	}

	// Delete head
	if !om.Delete("a") {
		t.Fatalf("expected delete a to return true")
	}
	if om.Front().Key != "c" {
		t.Fatalf("expected new head to be 'c', got '%s'", om.Front().Key)
	}

	// Delete tail
	if !om.Delete("d") {
		t.Fatalf("expected delete d to return true")
	}
	if om.Back().Key != "c" {
		t.Fatalf("expected new tail to be 'c', got '%s'", om.Back().Key)
	}

	// Delete last remaining
	if !om.Delete("c") {
		t.Fatalf("expected delete c to return true")
	}
	if om.Len() != 0 || om.Front() != nil || om.Back() != nil {
		t.Fatalf("expected empty map, got len %d", om.Len())
	}
}

func TestOrderedMap_ForEachAndEntries(t *testing.T) {
	om := New[int, string]()
	for i := 1; i <= 5; i++ {
		om.Set(i, fmt.Sprintf("val-%d", i))
	}

	var collected []int
	om.ForEach(func(key int, value string) bool {
		collected = append(collected, key)
		return key < 3 // early break after 3
	})

	if len(collected) != 3 {
		t.Fatalf("expected early break at 3, got %d elements", len(collected))
	}

	entries := om.Entries()
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}
	if entries[4].Value != "val-5" {
		t.Fatalf("expected entries[4] to be val-5, got %s", entries[4].Value)
	}
}

func TestOrderedMap_Sort(t *testing.T) {
	om := New[string, int]()
	om.Set("z", 1)
	om.Set("a", 20)
	om.Set("m", 5)

	om.Sort(func(a, b Entry[string, int]) bool {
		return a.Key < b.Key // sort alphabetically by key
	})

	keys := om.Keys()
	expected := []string{"a", "m", "z"}
	for i, k := range keys {
		if k != expected[i] {
			t.Fatalf("expected %s at index %d, got %s", expected[i], i, k)
		}
	}
}

func TestOrderedMap_JSON(t *testing.T) {
	om := New[string, any]()
	om.Set("name", "Arjun")
	om.Set("age", 28)
	om.Set("city", "Bengaluru")
	om.Set("active", true)

	raw, err := json.Marshal(om)
	if err != nil {
		t.Fatalf("failed to marshal ordered map: %v", err)
	}

	expectedJSON := `{"name":"Arjun","age":28,"city":"Bengaluru","active":true}`
	if string(raw) != expectedJSON {
		t.Fatalf("expected JSON %s, got %s", expectedJSON, string(raw))
	}

	// Test Unmarshal
	unmarshaled := New[string, any]()
	if err := json.Unmarshal(raw, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	unmarshaledKeys := unmarshaled.Keys()
	expectedKeys := []string{"name", "age", "city", "active"}
	for i, k := range unmarshaledKeys {
		if k != expectedKeys[i] {
			t.Fatalf("unmarshal key mismatch at %d: expected %s, got %s", i, expectedKeys[i], k)
		}
	}
}

func TestOrderedMap_Concurrent(t *testing.T) {
	om := NewConcurrent[int, int]()
	var wg sync.WaitGroup

	// Concurrently write 100 items
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			om.Set(idx, idx*10)
			om.Get(idx)
			om.Has(idx)
		}(i)
	}

	wg.Wait()

	if om.Len() != 100 {
		t.Fatalf("expected len 100, got %d", om.Len())
	}
}
