package queryfilter

import (
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestQueryFilter_Success(t *testing.T) {
	parser := New(Config{
		AllowedFilters: map[string]bool{
			"status": true,
			"age":    true,
			"role":   true,
		},
		AllowedSorts: map[string]bool{
			"created_at": true,
			"name":       true,
		},
		DefaultLimit: 20,
		MaxLimit:     100,
	})

	vals := url.Values{
		"status":   {"active"},
		"age[gte]": {"18"},
		"role[in]": {"admin,superadmin"},
		"sort":     {"-created_at,name"},
		"page":     {"2"},
		"limit":    {"10"},
	}

	res, err := parser.Parse(vals, 1)
	if err != nil {
		t.Fatalf("unexpected Parse error: %v", err)
	}

	if res.Limit != 10 {
		t.Fatalf("expected limit 10, got %d", res.Limit)
	}
	if res.Offset != 10 { // page 2 with limit 10 -> offset 10
		t.Fatalf("expected offset 10, got %d", res.Offset)
	}

	if !strings.Contains(res.OrderByClause, "created_at DESC") || !strings.Contains(res.OrderByClause, "name ASC") {
		t.Fatalf("unexpected order by: %s", res.OrderByClause)
	}

	if !strings.Contains(res.WhereClause, "status =") || !strings.Contains(res.WhereClause, "age >=") || !strings.Contains(res.WhereClause, "role IN") {
		t.Fatalf("unexpected where clause: %s", res.WhereClause)
	}

	// 1 (status) + 1 (age) + 2 (role in) = 4 args
	if len(res.Args) != 4 {
		t.Fatalf("expected 4 args, got %d", len(res.Args))
	}
}

func TestQueryFilter_DisallowedField(t *testing.T) {
	parser := New(Config{
		AllowedFilters: map[string]bool{"status": true},
	})

	// SQL Injection / Unwhitelisted attempt
	vals := url.Values{
		"password_hash": {"xyz"},
	}

	_, err := parser.Parse(vals, 1)
	if !errors.Is(err, ErrDisallowedField) {
		t.Fatalf("expected ErrDisallowedField, got: %v", err)
	}
}

func TestQueryFilter_DisallowedSort(t *testing.T) {
	parser := New(Config{
		AllowedSorts: map[string]bool{"created_at": true},
	})

	vals := url.Values{
		"sort": {"secret_column"},
	}

	_, err := parser.Parse(vals, 1)
	if !errors.Is(err, ErrDisallowedField) {
		t.Fatalf("expected ErrDisallowedField on sort, got: %v", err)
	}
}

func TestQueryFilter_InvalidOperator(t *testing.T) {
	parser := New(Config{
		AllowedFilters: map[string]bool{"age": true},
	})

	vals := url.Values{
		"age[bad_op]": {"18"},
	}

	_, err := parser.Parse(vals, 1)
	if !errors.Is(err, ErrInvalidOperator) {
		t.Fatalf("expected ErrInvalidOperator, got: %v", err)
	}
}
