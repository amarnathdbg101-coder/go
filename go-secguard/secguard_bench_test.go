package secguard

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkSecGuard_InjectSecurityHeaders(b *testing.B) {
	eng := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		eng.InjectSecurityHeaders(rec, req)
	}
}

func BenchmarkSecGuard_InspectWAF_CleanRequest(b *testing.B) {
	eng := New()
	path := "/api/v1/products/list"
	query := "category=electronics&page=2&limit=50&sort=desc"
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = eng.InspectWAF(path, query)
	}
}

func BenchmarkSecGuard_FullMiddleware(b *testing.B) {
	eng := New()
	defer eng.Close()

	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile?id=123", nil)
	rec := httptest.NewRecorder()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rec, req)
	}
}
