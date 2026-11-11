package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/shopsilo/go-network/ratelimit"
)

func TestLimiter_BasicAllow(t *testing.T) {
	limiter, err := ratelimit.New(
		ratelimit.NewMemoryStore(),
		ratelimit.Rule{Capacity: 3, Window: 1 * time.Minute},
	)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}
	defer limiter.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "192.168.1.50:1234"

	res1, err := limiter.Take(req)
	if err != nil || !res1.Allowed || res1.Remaining != 2 {
		t.Errorf("req 1: expected allowed=true, remaining=2; got allowed=%v, remaining=%d, err=%v", res1.Allowed, res1.Remaining, err)
	}

	res2, _ := limiter.Take(req)
	if !res2.Allowed || res2.Remaining != 1 {
		t.Errorf("req 2: expected allowed=true, remaining=1; got allowed=%v, remaining=%d", res2.Allowed, res2.Remaining)
	}

	res3, _ := limiter.Take(req)
	if !res3.Allowed || res3.Remaining != 0 {
		t.Errorf("req 3: expected allowed=true, remaining=0; got allowed=%v, remaining=%d", res3.Allowed, res3.Remaining)
	}

	res4, _ := limiter.Take(req)
	if res4.Allowed {
		t.Errorf("req 4: expected allowed=false for 4th request when capacity is 3")
	}
}

func TestLimiter_Middleware_Headers(t *testing.T) {
	limiter, _ := ratelimit.New(
		ratelimit.NewMemoryStore(),
		ratelimit.Rule{Capacity: 2, Window: 1 * time.Minute},
	)
	defer limiter.Close()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "10.0.0.1:5555"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w1.Code)
	}
	if w1.Header().Get("X-RateLimit-Limit") != "2" {
		t.Errorf("expected X-RateLimit-Limit=2, got %s", w1.Header().Get("X-RateLimit-Limit"))
	}

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.1:5555"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.RemoteAddr = "10.0.0.1:5555"
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", w3.Code)
	}
}

func TestKeyExtractors(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18")
	if ip := ratelimit.KeyByIP(req1); ip != "203.0.113.195" {
		t.Errorf("expected '203.0.113.195', got '%s'", ip)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-API-Key", "secret-key-123")
	keyFn := ratelimit.KeyByHeader("X-API-Key")
	if key := keyFn(req2); key != "X-API-Key:secret-key-123" {
		t.Errorf("expected 'X-API-Key:secret-key-123', got '%s'", key)
	}
}

func TestMemoryStore_Concurrent(t *testing.T) {
	limiter, _ := ratelimit.New(
		ratelimit.NewMemoryStore(),
		ratelimit.Rule{Capacity: 1000, Window: 1 * time.Minute},
	)
	defer limiter.Close()

	var wg sync.WaitGroup
	workers := 50
	requestsPerWorker := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "172.16.0.1:1234"
			for j := 0; j < requestsPerWorker; j++ {
				_, _ = limiter.Take(req)
			}
		}(i)
	}

	wg.Wait()

	reqFinal := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqFinal.RemoteAddr = "172.16.0.1:1234"
	res, err := limiter.Take(reqFinal)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Allowed {
		t.Errorf("expected 1001st request to be blocked")
	}
}
