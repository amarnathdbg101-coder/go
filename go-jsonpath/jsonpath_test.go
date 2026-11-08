package jsonpath

import (
	"errors"
	"testing"
)

const sampleJSON = `{
	"store": {
		"name": "SuperMart",
		"isOpen": true,
		"rating": 4.85,
		"branchCount": 12,
		"books": [
			{"title": "Go in Action", "price": 450, "inStock": true},
			{"title": "Clean Code", "price": 600, "inStock": false},
			{"title": "Concurrency in Go", "price": 550, "inStock": true}
		],
		"owner": {
			"firstName": "Amit",
			"lastName": "Verma"
		}
	}
}`

func TestGet_BasicFields(t *testing.T) {
	raw := []byte(sampleJSON)

	// String
	name, err := GetString(raw, "$.store.name")
	if err != nil || name != "SuperMart" {
		t.Fatalf("expected SuperMart, got %s (err: %v)", name, err)
	}

	// Without leading $
	first, err := GetString(raw, "store.owner.firstName")
	if err != nil || first != "Amit" {
		t.Fatalf("expected Amit, got %s (err: %v)", first, err)
	}

	// Bool
	isOpen, err := GetBool(raw, "$.store.isOpen")
	if err != nil || !isOpen {
		t.Fatalf("expected true, got %v (err: %v)", isOpen, err)
	}

	// Int
	branches, err := GetInt(raw, "$.store.branchCount")
	if err != nil || branches != 12 {
		t.Fatalf("expected 12, got %d (err: %v)", branches, err)
	}

	// Float
	rating, err := GetFloat(raw, "$.store.rating")
	if err != nil || rating != 4.85 {
		t.Fatalf("expected 4.85, got %f (err: %v)", rating, err)
	}
}

func TestGet_ArrayIndexing(t *testing.T) {
	raw := []byte(sampleJSON)

	title, err := GetString(raw, "$.store.books[1].title")
	if err != nil || title != "Clean Code" {
		t.Fatalf("expected 'Clean Code', got '%s' (err: %v)", title, err)
	}

	price, err := GetInt(raw, "$.store.books[0].price")
	if err != nil || price != 450 {
		t.Fatalf("expected 450, got %d (err: %v)", price, err)
	}
}

func TestGet_WildcardProjection(t *testing.T) {
	raw := []byte(sampleJSON)

	prices, err := GetSlice(raw, "$.store.books[*].price")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prices) != 3 {
		t.Fatalf("expected 3 prices, got %d", len(prices))
	}

	// Check values
	if prices[0].(float64) != 450 || prices[1].(float64) != 600 || prices[2].(float64) != 550 {
		t.Errorf("prices mismatch: %v", prices)
	}
}

func TestGet_Errors(t *testing.T) {
	raw := []byte(sampleJSON)

	// Not found
	_, err := Get(raw, "$.store.nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}

	// Index out of bounds
	_, err = Get(raw, "$.store.books[99].title")
	if !errors.Is(err, ErrIndexOutOfBounds) {
		t.Errorf("expected ErrIndexOutOfBounds, got: %v", err)
	}

	// Invalid syntax (unclosed bracket)
	_, err = Get(raw, "$.store.books[0")
	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath, got: %v", err)
	}

	// Invalid JSON
	_, err = Get([]byte("{invalid"), "$.store")
	if !errors.Is(err, ErrInvalidJSON) {
		t.Errorf("expected ErrInvalidJSON, got: %v", err)
	}

	// Type mismatch
	_, err = GetString(raw, "$.store.isOpen")
	if !errors.Is(err, ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch when reading bool as string, got: %v", err)
	}
}
