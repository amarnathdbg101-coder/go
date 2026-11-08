package gotree

import (
	"errors"
	"testing"
)

func TestTrie_InsertAndSearch(t *testing.T) {
	trie := NewTrie[int]()

	trie.Insert("apple", 100)
	trie.Insert("app", 50)
	trie.Insert("application", 200)
	trie.Insert("banana", 30)

	// Search exact
	val, ok := trie.Search("app")
	if !ok || val != 50 {
		t.Fatalf("expected 50 for 'app', got %d (ok=%v)", val, ok)
	}

	val, ok = trie.Search("apple")
	if !ok || val != 100 {
		t.Fatalf("expected 100 for 'apple', got %d", val)
	}

	// Missing
	_, ok = trie.Search("appl")
	if ok {
		t.Errorf("expected 'appl' to be missing")
	}

	// StartsWith prefix search
	matches := trie.StartsWith("app")
	if len(matches) != 3 {
		t.Errorf("expected 3 prefix matches for 'app', got %d", len(matches))
	}

	// Delete
	deleted := trie.Delete("app")
	if !deleted {
		t.Errorf("expected delete to return true")
	}

	_, ok = trie.Search("app")
	if ok {
		t.Errorf("expected 'app' to be deleted")
	}

	// 'apple' and 'application' should still exist
	val, ok = trie.Search("apple")
	if !ok || val != 100 {
		t.Errorf("expected 'apple' to remain intact")
	}
}

func TestHierarchy_CategoryTree(t *testing.T) {
	h := NewHierarchy[string]()

	// Root
	_ = h.Add("cat_electronics", "", "Electronics")

	// Level 1
	_ = h.Add("cat_mobiles", "cat_electronics", "Mobile Phones")
	_ = h.Add("cat_laptops", "cat_electronics", "Laptops")

	// Level 2
	_ = h.Add("cat_smartphones", "cat_mobiles", "Smartphones")

	// Test Ancestors (Breadcrumbs)
	ancestors, err := h.Ancestors("cat_smartphones")
	if err != nil {
		t.Fatalf("Ancestors failed: %v", err)
	}

	if len(ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(ancestors))
	}
	if ancestors[0].Value != "Electronics" || ancestors[1].Value != "Mobile Phones" {
		t.Errorf("breadcrumb mismatch: %v -> %v", ancestors[0].Value, ancestors[1].Value)
	}

	// Test Descendants
	descendants, err := h.Descendants("cat_electronics")
	if err != nil {
		t.Fatalf("Descendants failed: %v", err)
	}
	if len(descendants) != 3 {
		t.Errorf("expected 3 descendants under electronics, got %d", len(descendants))
	}

	// Duplicate node error
	err = h.Add("cat_laptops", "", "Duplicate")
	if !errors.Is(err, ErrDuplicateNode) {
		t.Errorf("expected ErrDuplicateNode, got %v", err)
	}

	// Parent not found error
	err = h.Add("cat_unknown", "cat_ghost", "Ghost Child")
	if !errors.Is(err, ErrParentNotFound) {
		t.Errorf("expected ErrParentNotFound, got %v", err)
	}
}
