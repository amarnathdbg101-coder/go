package retryablehttp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_SuccessWithoutRetry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := New(WithMaxRetries(3))
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestClient_RetryOnServerErrorAndRecover(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		curr := atomic.AddInt32(&attempts, 1)
		if curr < 3 {
			w.WriteHeader(http.StatusBadGateway) // 502
			return
		}
		w.WriteHeader(http.StatusOK) // 3rd attempt succeeds
		_, _ = w.Write([]byte("recovered"))
	}))
	defer server.Close()

	var retryEvents int32
	client := New(
		WithMaxRetries(4),
		WithWaitLimits(5*time.Millisecond, 20*time.Millisecond),
		WithOnRetry(func(attempt int, req *http.Request, resp *http.Response, err error) {
			atomic.AddInt32(&retryEvents, 1)
		}),
	)

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected recovery, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Fatalf("expected 3 server hits, got %d", attempts)
	}
	if atomic.LoadInt32(&retryEvents) != 2 {
		t.Fatalf("expected 2 retry events, got %d", retryEvents)
	}
}

func TestClient_PayloadRewind(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		expected := `{"msg":"hello world"}`
		if string(body) != expected {
			t.Errorf("server received mangled body: %s", string(body))
		}

		curr := atomic.AddInt32(&attempts, 1)
		if curr == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(
		WithMaxRetries(2),
		WithWaitLimits(5*time.Millisecond, 15*time.Millisecond),
	)

	req, err := NewRequest(context.Background(), http.MethodPost, server.URL, strings.NewReader(`{"msg":"hello world"}`))
	if err != nil {
		t.Fatalf("failed to create req: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to do req: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestClient_RetryAfterHeader(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		curr := atomic.AddInt32(&attempts, 1)
		if curr == 1 {
			w.Header().Set("Retry-After", "1") // wait 1 sec
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(
		WithMaxRetries(2),
		WithWaitLimits(5*time.Millisecond, 20*time.Millisecond),
	)

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if elapsed < 900*time.Millisecond {
		t.Fatalf("expected sleep of ~1s due to Retry-After, elapsed: %v", elapsed)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := New(
		WithMaxRetries(5),
		WithWaitLimits(500*time.Millisecond, 1*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	_, err := client.Do(req)
	if err == nil || !strings.Contains(err.Error(), "context") {
		t.Fatalf("expected context deadline exceeded error, got: %v", err)
	}
}
