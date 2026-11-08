package urlquery

import (
	"errors"
	"net/url"
	"testing"
	"time"
)

type FilterQuery struct {
	Page    int           `query:"page,default=1"`
	Limit   int           `query:"limit,default=20"`
	Search  string        `query:"q"`
	Active  bool          `query:"active,default=true"`
	Rating  float64       `query:"min_rating,default=4.5"`
	Timeout time.Duration `query:"timeout,default=5s"`
	Tags    []string      `query:"tags,comma"`
	IDs     []int         `query:"ids"`
	Secret  string        `query:"-"`
}

func TestDecode_Values(t *testing.T) {
	vals := url.Values{
		"page":       []string{"2"},
		"q":          []string{"golang tutorial"},
		"min_rating": []string{"4.8"},
		"timeout":    []string{"10s"},
		"tags":       []string{"go, backend, grpc"},
		"ids":        []string{"101", "102", "103"},
	}

	var q FilterQuery
	if err := Decode(vals, &q); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if q.Page != 2 {
		t.Errorf("expected page 2, got %d", q.Page)
	}
	if q.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", q.Limit)
	}
	if q.Search != "golang tutorial" {
		t.Errorf("expected search 'golang tutorial', got '%s'", q.Search)
	}
	if !q.Active {
		t.Errorf("expected default active true")
	}
	if q.Rating != 4.8 {
		t.Errorf("expected rating 4.8, got %f", q.Rating)
	}
	if q.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", q.Timeout)
	}
	if len(q.Tags) != 3 || q.Tags[1] != "backend" {
		t.Errorf("tags mismatch: %v", q.Tags)
	}
	if len(q.IDs) != 3 || q.IDs[2] != 103 {
		t.Errorf("IDs mismatch: %v", q.IDs)
	}
}

func TestDecodeURL(t *testing.T) {
	rawURL := "https://api.example.com/products?page=3&limit=50&active=false"

	var q FilterQuery
	if err := DecodeURL(rawURL, &q); err != nil {
		t.Fatalf("decodeURL failed: %v", err)
	}

	if q.Page != 3 || q.Limit != 50 || q.Active {
		t.Errorf("mismatch decodeURL: %+v", q)
	}
}

func TestEncode(t *testing.T) {
	q := FilterQuery{
		Page:   1,
		Limit:  10,
		Search: "laptop",
		Active: true,
		Tags:   []string{"tech", "apple"},
	}

	encodedStr, err := EncodeString(q)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	vals, _ := url.ParseQuery(encodedStr)
	if vals.Get("page") != "1" || vals.Get("q") != "laptop" || vals.Get("tags") != "tech,apple" {
		t.Errorf("encoded query mismatch: %s", encodedStr)
	}
}

func TestErrors(t *testing.T) {
	// Not a pointer
	var q FilterQuery
	if !errors.Is(Decode(url.Values{}, q), ErrNotPointerToStruct) {
		t.Errorf("expected ErrNotPointerToStruct")
	}

	// Invalid integer
	vals := url.Values{"page": []string{"not-a-number"}}
	if !errors.Is(Decode(vals, &q), ErrInvalidValue) {
		t.Errorf("expected ErrInvalidValue for bad int")
	}
}
