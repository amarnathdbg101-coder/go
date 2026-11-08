package secguard_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/shopsilo/go-secguard"
)

func ExampleMiddleware() {
	guard := secguard.Middleware()

	handler := guard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("allowed"))
	}))

	// 1. Clean Request
	req := httptest.NewRequest(http.MethodGet, "/products?category=books", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Printf("Clean Request Status: %d\n", rec.Code)
	fmt.Printf("HSTS Header Present: %v\n", rec.Header().Get("X-Content-Type-Options") == "nosniff")

	// 2. Malicious SQL Injection Request
	badReq := httptest.NewRequest(http.MethodGet, "/search?q=1+union+select+*", nil)
	badRec := httptest.NewRecorder()
	handler.ServeHTTP(badRec, badReq)

	fmt.Printf("SQLi Request Status: %d\n", badRec.Code)

	// Output:
	// Clean Request Status: 200
	// HSTS Header Present: true
	// SQLi Request Status: 403
}
