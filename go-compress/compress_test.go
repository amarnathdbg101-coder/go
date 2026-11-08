package gocompress

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzip_Roundtrip(t *testing.T) {
	original := []byte("ShopSilo High-Performance Microservice Architecture 2026")

	compressed, err := Gzip(original)
	if err != nil {
		t.Fatalf("Gzip failed: %v", err)
	}

	decompressed, err := Gunzip(compressed)
	if err != nil {
		t.Fatalf("Gunzip failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Fatalf("roundtrip mismatch: expected %s, got %s", original, decompressed)
	}
}

func TestMiddleware_CompressesLargePayload(t *testing.T) {
	comp := New(Config{MinSize: 100}) // Lower threshold for test

	largeText := strings.Repeat("Product Details JSON Payload, ", 50) // ~1.5 KB

	target := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(largeText))
	})

	handler := comp.Handler(target)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding: gzip, got %q", rec.Header().Get("Content-Encoding"))
	}

	// Decompress response body to verify integrity
	decompressed, err := Gunzip(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("failed to decompress gzip response: %v", err)
	}

	if string(decompressed) != largeText {
		t.Fatal("decompressed response does not match original payload")
	}
}

func TestMiddleware_SkipsSmallPayload(t *testing.T) {
	comp := New(Config{MinSize: 1024}) // 1KB threshold

	smallText := "{\"status\":\"ok\"}" // ~15 bytes

	target := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(smallText))
	})

	handler := comp.Handler(target)

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("payload below threshold should NOT be compressed")
	}

	if rec.Body.String() != smallText {
		t.Fatalf("expected raw body %q, got %q", smallText, rec.Body.String())
	}
}

func TestMiddleware_NoAcceptEncoding(t *testing.T) {
	comp := New(Config{MinSize: 10})

	payload := strings.Repeat("hello world ", 20)

	target := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	})

	handler := comp.Handler(target)

	req := httptest.NewRequest(http.MethodGet, "/data", nil) // No Accept-Encoding header
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("should not compress when client doesn't support gzip")
	}
}
