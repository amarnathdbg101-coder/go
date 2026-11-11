package paginate_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-paginate"
)

// -------------------------------------------------------------------------
// 1. Parameter Extraction & Defaults Tests
// -------------------------------------------------------------------------

func TestFromRequest_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	params := paginate.FromRequest(req)

	if params.Page != 1 {
		t.Errorf("expected default page 1, got %d", params.Page)
	}
	if params.PerPage != 10 {
		t.Errorf("expected default per_page 10, got %d", params.PerPage)
	}
	if params.Sort != "id" {
		t.Errorf("expected default sort 'id', got '%s'", params.Sort)
	}
	if params.Order != "DESC" {
		t.Errorf("expected default order 'DESC', got '%s'", params.Order)
	}
	if params.Offset() != 0 {
		t.Errorf("expected offset 0 for page 1, got %d", params.Offset())
	}
}

func TestFromRequest_CustomValues_And_Bounds(t *testing.T) {
	// Case A: Custom query parameters
	req := httptest.NewRequest(http.MethodGet, "/products?page=3&per_page=25&sort=price&order=asc", nil)
	params := paginate.FromRequest(req)

	if params.Page != 3 {
		t.Errorf("expected page 3, got %d", params.Page)
	}
	if params.PerPage != 25 {
		t.Errorf("expected per_page 25, got %d", params.PerPage)
	}
	if params.Offset() != 50 { // (3 - 1) * 25 = 50
		t.Errorf("expected offset 50, got %d", params.Offset())
	}
	if params.Sort != "price" {
		t.Errorf("expected sort 'price', got '%s'", params.Sort)
	}
	if params.Order != "ASC" {
		t.Errorf("expected order 'ASC', got '%s'", params.Order)
	}

	// Case B: Boundary limits (max per_page capping)
	reqFlood := httptest.NewRequest(http.MethodGet, "/products?per_page=5000", nil)
	paramsFlood := paginate.FromRequest(reqFlood)
	if paramsFlood.PerPage != paginate.MaxPerPage {
		t.Errorf("expected per_page capped at %d, got %d", paginate.MaxPerPage, paramsFlood.PerPage)
	}

	// Case C: Negative or invalid values fall back to defaults
	reqBad := httptest.NewRequest(http.MethodGet, "/products?page=-5&per_page=-20&order=invalid", nil)
	paramsBad := paginate.FromRequest(reqBad)
	if paramsBad.Page != 1 || paramsBad.PerPage != 10 || paramsBad.Order != "DESC" {
		t.Errorf("expected fallback to defaults for bad inputs, got: %+v", paramsBad)
	}
}

func TestFromRequest_AlternativeAliases(t *testing.T) {
	// Tests limit, sort_by, and dir aliases
	req := httptest.NewRequest(http.MethodGet, "/items?limit=15&sort_by=name&dir=asc", nil)
	params := paginate.FromRequest(req)

	if params.PerPage != 15 || params.Sort != "name" || params.Order != "ASC" {
		t.Errorf("expected aliases to be parsed, got: %+v", params)
	}
}

// -------------------------------------------------------------------------
// 2. SQL Injection Defense & Whitelist Tests
// -------------------------------------------------------------------------

func TestSanitizeSort_SQLInjectionProtection(t *testing.T) {
	allowed := []string{"id", "created_at", "price", "title"}

	// Attack 1: Classic semicolon drop table attack
	cleanSort, cleanOrder := paginate.SanitizeSort("price; DROP TABLE users;--", "asc", "id", allowed...)
	if cleanSort != "id" {
		t.Errorf("expected fallback to 'id' on SQL injection attempt, got '%s'", cleanSort)
	}
	if cleanOrder != "ASC" {
		t.Errorf("expected valid order 'ASC', got '%s'", cleanOrder)
	}

	// Attack 2: Order by sleep injection
	_, badOrder := paginate.SanitizeSort("created_at", "DESC; SELECT pg_sleep(5);--", "id", allowed...)
	if badOrder != "DESC" {
		t.Errorf("expected order strictly fallback to 'DESC', got '%s'", badOrder)
	}

	// Case 3: Whitelisted column matches successfully (case insensitive)
	validSort, validOrder := paginate.SanitizeSort("PRICE", "asc", "id", allowed...)
	if validSort != "price" || validOrder != "ASC" {
		t.Errorf("expected 'price' and 'ASC', got '%s' and '%s'", validSort, validOrder)
	}

	// Case 4: Params.Sanitize method
	p := paginate.Params{Page: 1, PerPage: 10, Sort: "unknown_hack_col", Order: "desc"}
	cleanP := p.Sanitize("created_at", allowed...)
	if cleanP.Sort != "created_at" {
		t.Errorf("expected default 'created_at', got '%s'", cleanP.Sort)
	}
}

// -------------------------------------------------------------------------
// 3. Metadata Calculation Math Tests
// -------------------------------------------------------------------------

func TestNewMeta_Calculations(t *testing.T) {
	// Case A: 0 total records
	meta0 := paginate.NewMeta(paginate.Params{Page: 1, PerPage: 10}, 0)
	if meta0.TotalPages != 0 || meta0.HasNext || meta0.HasPrev {
		t.Errorf("expected empty meta for 0 records, got: %+v", meta0)
	}

	// Case B: 95 records with per_page = 10 -> 10 total pages
	p1 := paginate.Params{Page: 1, PerPage: 10}
	meta1 := paginate.NewMeta(p1, 95)
	if meta1.TotalPages != 10 {
		t.Errorf("expected 10 total pages, got %d", meta1.TotalPages)
	}
	if !meta1.HasNext || meta1.HasPrev {
		t.Errorf("page 1 should have next=true and prev=false, got next=%v prev=%v", meta1.HasNext, meta1.HasPrev)
	}

	// Middle page: Page 5
	p5 := paginate.Params{Page: 5, PerPage: 10}
	meta5 := paginate.NewMeta(p5, 95)
	if !meta5.HasNext || !meta5.HasPrev {
		t.Errorf("page 5 should have next=true and prev=true")
	}

	// Last page: Page 10
	p10 := paginate.Params{Page: 10, PerPage: 10}
	meta10 := paginate.NewMeta(p10, 95)
	if meta10.HasNext || !meta10.HasPrev {
		t.Errorf("last page should have next=false and prev=true")
	}

	// Case C: Exact multiple (100 records / 20 per_page = 5 pages)
	pExact := paginate.Params{Page: 5, PerPage: 20}
	metaExact := paginate.NewMeta(pExact, 100)
	if metaExact.TotalPages != 5 {
		t.Errorf("expected exact 5 pages, got %d", metaExact.TotalPages)
	}
}

// -------------------------------------------------------------------------
// 4. Keyset Cursor Pagination Tests
// -------------------------------------------------------------------------

func TestCursor_SimpleEncodeDecode(t *testing.T) {
	originalID := "prod_abc_9999"
	cursor := paginate.EncodeCursor(originalID)

	if cursor == "" {
		t.Fatalf("expected non-empty base64 cursor")
	}

	decodedID, err := paginate.DecodeCursor(cursor)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decodedID != originalID {
		t.Errorf("expected ID '%s', got '%s'", originalID, decodedID)
	}

	// Empty cursor decode returns empty string without error
	emptyDecoded, err := paginate.DecodeCursor("")
	if err != nil || emptyDecoded != "" {
		t.Errorf("empty cursor decode should return empty string")
	}

	// Corrupted cursor returns ErrInvalidCursor
	_, errBad := paginate.DecodeCursor("!!not_base_64!!")
	if errBad != paginate.ErrInvalidCursor {
		t.Errorf("expected ErrInvalidCursor on corrupted input, got %v", errBad)
	}
}

func TestCursor_CompositeEncodeDecode(t *testing.T) {
	now := time.Now().Truncate(time.Nanosecond)
	originalID := "order_789"

	cursor := paginate.EncodeCompositeCursor(now, originalID)
	decodedTime, decodedID, err := paginate.DecodeCompositeCursor(cursor)

	if err != nil {
		t.Fatalf("composite decode error: %v", err)
	}
	if decodedID != originalID {
		t.Errorf("expected ID '%s', got '%s'", originalID, decodedID)
	}
	if decodedTime.UnixNano() != now.UnixNano() {
		t.Errorf("expected timestamp %v, got %v", now.UnixNano(), decodedTime.UnixNano())
	}
}

// -------------------------------------------------------------------------
// 5. PostgreSQL Query Builder Tests
// -------------------------------------------------------------------------

func TestPostgres_SQLClause(t *testing.T) {
	p := paginate.Params{
		Page:    2,
		PerPage: 15,
		Sort:    "created_at",
		Order:   "DESC",
	}

	// Case 1: Start index 1
	clause1, args1 := paginate.SQLClause(p, 1)
	expectedClause1 := `ORDER BY "created_at" DESC LIMIT $1 OFFSET $2`
	if clause1 != expectedClause1 {
		t.Errorf("expected clause '%s', got '%s'", expectedClause1, clause1)
	}
	if len(args1) != 2 || args1[0] != 15 || args1[1] != 15 {
		t.Errorf("expected args [15, 15], got %+v", args1)
	}

	// Case 2: Start index 3 (after WHERE status = $1 AND category = $2)
	clause3, args3 := paginate.SQLClause(p, 3)
	expectedClause3 := `ORDER BY "created_at" DESC LIMIT $3 OFFSET $4`
	if clause3 != expectedClause3 {
		t.Errorf("expected clause '%s', got '%s'", expectedClause3, clause3)
	}
	if len(args3) != 2 || args3[0] != 15 || args3[1] != 15 {
		t.Errorf("expected args [15, 15], got %+v", args3)
	}

	// Case 3: Table qualified identifier "products.price"
	pTable := paginate.Params{Page: 1, PerPage: 10, Sort: "products.price", Order: "ASC"}
	clauseTable, _ := paginate.SQLClause(pTable, 1)
	if !strings.Contains(clauseTable, `"products"."price" ASC`) {
		t.Errorf("expected table-qualified identifier, got: %s", clauseTable)
	}
}

func TestPostgres_BuildCountSQL(t *testing.T) {
	baseSQL := "SELECT p.id, p.name FROM products p WHERE p.is_active = $1"
	countSQL := paginate.BuildCountSQL(baseSQL)

	expected := "SELECT COUNT(*) FROM (SELECT p.id, p.name FROM products p WHERE p.is_active = $1) AS _count_subquery"
	if countSQL != expected {
		t.Errorf("expected count query '%s', got '%s'", expected, countSQL)
	}
}

func TestPostgres_CursorWhereClause(t *testing.T) {
	whereClause, orderClause, args := paginate.CursorWhereClause("id", "prod_100", "DESC", 1)

	if whereClause != `"id" < $1` {
		t.Errorf("expected where clause '\"id\" < $1', got '%s'", whereClause)
	}
	if orderClause != `ORDER BY "id" DESC` {
		t.Errorf("expected order clause, got '%s'", orderClause)
	}
	if len(args) != 1 || args[0] != "prod_100" {
		t.Errorf("expected arg 'prod_100', got %+v", args)
	}
}

// -------------------------------------------------------------------------
// 6. In-Memory Slice Pagination & JSON Serialization Tests
// -------------------------------------------------------------------------

func TestPaginateSlice(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	params := paginate.Params{Page: 2, PerPage: 4}

	result := paginate.PaginateSlice(items, params)

	if len(result.Data) != 4 {
		t.Fatalf("expected 4 items on page 2, got %d", len(result.Data))
	}
	if result.Data[0] != "e" || result.Data[3] != "h" {
		t.Errorf("expected items [e, f, g, h], got %+v", result.Data)
	}
	if result.Meta.TotalRecords != 10 || result.Meta.TotalPages != 3 {
		t.Errorf("meta mismatch: %+v", result.Meta)
	}
}

func TestJSON_Serialization_NullArrayPrevention(t *testing.T) {
	meta := paginate.NewMeta(paginate.Params{Page: 1, PerPage: 10}, 0)
	page := paginate.NewPage[string](nil, meta)

	bytes, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}

	jsonStr := string(bytes)
	// Must contain `"data":[]` rather than `"data":null` for frontend friendliness
	if !strings.Contains(jsonStr, `"data":[]`) {
		t.Errorf("expected empty array [] in JSON, got: %s", jsonStr)
	}
}

// -------------------------------------------------------------------------
// 7. Keyset Cursor Request Extraction & Options Tests
// -------------------------------------------------------------------------

func TestCursorFromRequest_And_Values(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/feed?limit=30&cursor=ZXhhbXBsZQ==&order=asc", nil)
	cp := paginate.CursorFromRequest(req, 20)

	if cp.Limit != 30 {
		t.Errorf("expected limit 30, got %d", cp.Limit)
	}
	if cp.Cursor != "ZXhhbXBsZQ==" {
		t.Errorf("expected cursor 'ZXhhbXBsZQ==', got '%s'", cp.Cursor)
	}
	if cp.Order != "ASC" {
		t.Errorf("expected order 'ASC', got '%s'", cp.Order)
	}

	// Default fallback
	reqEmpty := httptest.NewRequest(http.MethodGet, "/feed", nil)
	cpDefault := paginate.CursorFromRequest(reqEmpty, 15)
	if cpDefault.Limit != 15 || cpDefault.Cursor != "" || cpDefault.Order != "DESC" {
		t.Errorf("expected default values, got: %+v", cpDefault)
	}
}

func TestConfig_Options(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users?per_page=500", nil)
	params := paginate.FromRequest(req,
		paginate.WithDefaults(2, 25, "created_at", "ASC"),
		paginate.WithMaxPerPage(200),
	)

	// Since per_page=500 exceeds max 200, it should be capped at 200
	if params.PerPage != 200 {
		t.Errorf("expected per_page capped at 200, got %d", params.PerPage)
	}
}

func TestPostgres_SQLClauseWithOptions(t *testing.T) {
	params := paginate.Params{
		Page:    1,
		PerPage: 10,
		Sort:    "deleted_at",
		Order:   "DESC",
	}

	// Test NULLS LAST with tie breaker
	clause, args := paginate.SQLClauseWithOptions(params, 1, paginate.NullsLast, "id")

	expectedClause := `ORDER BY "deleted_at" DESC NULLS LAST, "id" DESC LIMIT $1 OFFSET $2`
	if clause != expectedClause {
		t.Errorf("expected '%s', got '%s'", expectedClause, clause)
	}
	if len(args) != 2 || args[0] != 10 || args[1] != 0 {
		t.Errorf("unexpected args: %+v", args)
	}
}

func TestPostgres_CompositeCursorWhereClause(t *testing.T) {
	createdAt := time.Now()
	id := "order_99"

	whereClause, orderClause, args := paginate.CompositeCursorWhereClause("created_at", "id", createdAt, id, "DESC", 1)

	expectedWhere := `("created_at", "id") < ($1, $2)`
	expectedOrder := `ORDER BY "created_at" DESC, "id" DESC`

	if whereClause != expectedWhere {
		t.Errorf("expected where '%s', got '%s'", expectedWhere, whereClause)
	}
	if orderClause != expectedOrder {
		t.Errorf("expected order '%s', got '%s'", expectedOrder, orderClause)
	}
	if len(args) != 2 || args[0] != createdAt || args[1] != id {
		t.Errorf("unexpected args: %+v", args)
	}

	// Nil cursor values should return empty where clause
	emptyWhere, _, _ := paginate.CompositeCursorWhereClause("created_at", "id", nil, nil, "DESC", 1)
	if emptyWhere != "" {
		t.Errorf("expected empty where for nil values, got '%s'", emptyWhere)
	}
}

func TestBuildCursorPage(t *testing.T) {
	type Item struct {
		ID   string
		Name string
	}

	// Query fetched limit + 1 items (3 items for limit=2)
	items := []Item{
		{ID: "1", Name: "Alpha"},
		{ID: "2", Name: "Beta"},
		{ID: "3", Name: "Gamma"}, // Extra item indicating has_more=true
	}

	page := paginate.BuildCursorPage(items, 2, func(item Item) string {
		return paginate.EncodeCursor(item.ID)
	})

	if len(page.Data) != 2 {
		t.Fatalf("expected 2 items after trimming, got %d", len(page.Data))
	}
	if !page.Meta.HasMore {
		t.Errorf("expected HasMore to be true")
	}
	if page.Meta.Limit != 2 {
		t.Errorf("expected limit 2, got %d", page.Meta.Limit)
	}

	decodedID, err := paginate.DecodeCursor(page.Meta.NextCursor)
	if err != nil || decodedID != "2" {
		t.Fatalf("expected next cursor to point to item 2, got '%s', err: %v", decodedID, err)
	}

	// Test exact count (no more items)
	exactItems := []Item{
		{ID: "1", Name: "Alpha"},
	}
	pageExact := paginate.BuildCursorPage(exactItems, 2, func(item Item) string {
		return paginate.EncodeCursor(item.ID)
	})
	if pageExact.Meta.HasMore {
		t.Errorf("expected HasMore false for exact count")
	}
	if pageExact.Meta.NextCursor != "" {
		t.Errorf("expected empty NextCursor when no more pages, got '%s'", pageExact.Meta.NextCursor)
	}
}

func TestCursor_EdgeCases(t *testing.T) {
	// Empty cursor where clause
	where, order, args := paginate.CursorWhereClause("id", "", "ASC", 1)
	if where != "" || !strings.Contains(order, "ASC") || len(args) != 0 {
		t.Errorf("expected empty where for empty cursorVal, got '%s', '%s'", where, order)
	}

	// Limit fallback when PerPage <= 0
	p := paginate.Params{PerPage: 0}
	if p.Limit() != paginate.DefaultPerPage {
		t.Errorf("expected default limit %d, got %d", paginate.DefaultPerPage, p.Limit())
	}

	// DecodeCompositeCursor invalid cases
	_, _, err := paginate.DecodeCompositeCursor("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	// Valid base64 but missing separator ":"
	_, _, err = paginate.DecodeCompositeCursor("MTIzNDU=") // "12345"
	if err == nil {
		t.Error("expected error for composite cursor without colon")
	}

	// Valid base64 and colon but not integer nano
	_, _, err = paginate.DecodeCompositeCursor("bm90LWludDppZA==") // "not-int:id"
	if err == nil {
		t.Error("expected error for non-integer timestamp nano")
	}

	// Empty composite cursor
	ts, id, err := paginate.DecodeCompositeCursor("")
	if err != nil || !ts.IsZero() || id != "" {
		t.Errorf("expected zero values for empty composite cursor")
	}
}
