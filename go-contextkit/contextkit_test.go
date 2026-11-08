package gocontextkit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContextKit_Values(t *testing.T) {
	ctx := context.Background()

	ctx = WithRequestID(ctx, "req-12345")
	ctx = WithCorrelationID(ctx, "corr-99999")
	ctx = WithUserID(ctx, "usr-888")
	ctx = WithTenantID(ctx, "org-alpha")
	ctx = WithClientIP(ctx, "203.0.113.195")
	ctx = WithMetadata(ctx, "region", "ap-south-1")

	if GetRequestID(ctx) != "req-12345" {
		t.Fatalf("GetRequestID mismatch: %s", GetRequestID(ctx))
	}
	if GetCorrelationID(ctx) != "corr-99999" {
		t.Fatalf("GetCorrelationID mismatch: %s", GetCorrelationID(ctx))
	}
	if GetUserID(ctx) != "usr-888" {
		t.Fatalf("GetUserID mismatch: %s", GetUserID(ctx))
	}
	if GetTenantID(ctx) != "org-alpha" {
		t.Fatalf("GetTenantID mismatch: %s", GetTenantID(ctx))
	}
	if GetClientIP(ctx) != "203.0.113.195" {
		t.Fatalf("GetClientIP mismatch: %s", GetClientIP(ctx))
	}

	val, ok := GetMetadata(ctx, "region")
	if !ok || val != "ap-south-1" {
		t.Fatalf("GetMetadata mismatch: %s, ok: %v", val, ok)
	}
}

func TestContextKit_Detach(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	parent = WithUserID(parent, "usr-42")
	parent = WithRequestID(parent, "req-xyz")

	detached := Detach(parent)

	// Cancel parent
	cancel()

	// Parent should be canceled
	if parent.Err() != context.Canceled {
		t.Fatalf("expected parent to be canceled, got %v", parent.Err())
	}

	// Detached must NOT be canceled
	if detached.Err() != nil {
		t.Fatalf("detached context should not have error, got %v", detached.Err())
	}
	if detached.Done() != nil {
		t.Fatal("detached context Done() channel must be nil")
	}
	deadline, ok := detached.Deadline()
	if ok || !deadline.IsZero() {
		t.Fatal("detached context should have no deadline")
	}

	// Detached context MUST still retain parent values!
	if GetUserID(detached) != "usr-42" {
		t.Fatalf("expected usr-42 on detached context, got %s", GetUserID(detached))
	}
	if GetRequestID(detached) != "req-xyz" {
		t.Fatalf("expected req-xyz on detached context, got %s", GetRequestID(detached))
	}
}

func TestHTTPMiddleware(t *testing.T) {
	var capturedReqID string
	var capturedIP string

	target := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReqID = GetRequestID(r.Context())
		capturedIP = GetClientIP(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := HTTPMiddleware(target)

	// Case 1: Client provides X-Request-ID and X-Forwarded-For
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "custom-trace-id")
	req.Header.Set("X-Forwarded-For", "198.51.100.22, 10.0.0.1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedReqID != "custom-trace-id" {
		t.Fatalf("expected custom-trace-id, got %s", capturedReqID)
	}
	if capturedIP != "198.51.100.22" {
		t.Fatalf("expected 198.51.100.22, got %s", capturedIP)
	}
	if rec.Header().Get("X-Request-ID") != "custom-trace-id" {
		t.Fatalf("response header did not echo X-Request-ID: %s", rec.Header().Get("X-Request-ID"))
	}

	// Case 2: No X-Request-ID -> Middleware automatically generates one
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if capturedReqID == "" || len(capturedReqID) < 16 {
		t.Fatalf("expected auto-generated request ID, got %s", capturedReqID)
	}
}
