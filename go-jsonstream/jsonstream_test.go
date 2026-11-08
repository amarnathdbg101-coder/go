package jsonstream

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type Product struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Price int    `json:"price"`
}

func TestStreamArray(t *testing.T) {
	jsonPayload := `[
		{"id": 1, "title": "Keyboard", "price": 1200},
		{"id": 2, "title": "Mouse", "price": 600},
		{"id": 3, "title": "Monitor", "price": 9500}
	]`

	var products []Product
	err := StreamArray(strings.NewReader(jsonPayload), func(p Product) error {
		products = append(products, p)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(products) != 3 {
		t.Fatalf("expected 3 products, got %d", len(products))
	}
	if products[1].Title != "Mouse" || products[2].Price != 9500 {
		t.Errorf("mismatch product content: %+v", products)
	}
}

func TestStreamArrayChan(t *testing.T) {
	jsonPayload := `[
		{"id": 10, "title": "Laptop", "price": 55000},
		{"id": 20, "title": "Tablet", "price": 25000}
	]`

	ctx := context.Background()
	ch := StreamArrayChan[Product](ctx, strings.NewReader(jsonPayload), 2)

	var items []Product
	for res := range ch {
		if res.Err != nil {
			t.Fatalf("stream chan error: %v", res.Err)
		}
		items = append(items, res.Value)
	}

	if len(items) != 2 || items[0].ID != 10 {
		t.Errorf("expected 2 items, got %+v", items)
	}
}

func TestStreamNDJSON(t *testing.T) {
	ndjsonPayload := `{"id": 1, "title": "Pencil", "price": 10}
{"id": 2, "title": "Notebook", "price": 50}
{"id": 3, "title": "Backpack", "price": 800}`

	var count int
	err := StreamNDJSON(strings.NewReader(ndjsonPayload), func(p Product) error {
		count++
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 lines processed, got %d", count)
	}
}

func TestArrayWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewArrayWriter[Product](&buf)

	p1 := Product{ID: 1, Title: "Pen", Price: 20}
	p2 := Product{ID: 2, Title: "Eraser", Price: 5}

	if err := writer.Write(p1); err != nil {
		t.Fatalf("failed to write p1: %v", err)
	}
	if err := writer.Write(p2); err != nil {
		t.Fatalf("failed to write p2: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Verify the generated JSON array is streamable
	var decoded []Product
	err := StreamArray(&buf, func(p Product) error {
		decoded = append(decoded, p)
		return nil
	})

	if err != nil {
		t.Fatalf("failed to stream-read written array: %v", err)
	}
	if len(decoded) != 2 || decoded[0].Title != "Pen" || decoded[1].Price != 5 {
		t.Errorf("decoded array mismatch: %+v", decoded)
	}
}

func TestNDJSONWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewNDJSONWriter[Product](&buf)

	writer.Write(Product{ID: 100, Title: "Phone", Price: 15000})
	writer.Write(Product{ID: 101, Title: "Charger", Price: 999})

	var readItems []Product
	err := StreamNDJSON(&buf, func(p Product) error {
		readItems = append(readItems, p)
		return nil
	})

	if err != nil {
		t.Fatalf("stream ndjson read failed: %v", err)
	}
	if len(readItems) != 2 || readItems[1].Title != "Charger" {
		t.Errorf("mismatch read NDJSON items: %+v", readItems)
	}
}

func TestErrors(t *testing.T) {
	// Not an array
	notArray := `{"title": "SingleObject"}`
	err := StreamArray(strings.NewReader(notArray), func(p Product) error {
		return nil
	})
	if !errors.Is(err, ErrExpectedArray) {
		t.Errorf("expected ErrExpectedArray, got %v", err)
	}

	// Truncated JSON
	truncated := `[{"id": 1, "title": "Bro`
	err = StreamArray(strings.NewReader(truncated), func(p Product) error {
		return nil
	})
	if !errors.Is(err, ErrMalformedJSON) {
		t.Errorf("expected ErrMalformedJSON, got %v", err)
	}
}
