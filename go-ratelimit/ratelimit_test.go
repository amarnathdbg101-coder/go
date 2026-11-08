package ratelimit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/shopsilo/go-ratelimit"
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

	// Request 1: Allowed (Remaining = 2)
	res1, err := limiter.Take(req)
	if err != nil || !res1.Allowed || res1.Remaining != 2 {
		t.Errorf("req 1: expected allowed=true, remaining=2; got allowed=%v, remaining=%d, err=%v", res1.Allowed, res1.Remaining, err)
	}

	// Request 2: Allowed (Remaining = 1)
	res2, _ := limiter.Take(req)
	if !res2.Allowed || res2.Remaining != 1 {
		t.Errorf("req 2: expected allowed=true, remaining=1; got allowed=%v, remaining=%d", res2.Allowed, res2.Remaining)
	}

	// Request 3: Allowed (Remaining = 0)
	res3, _ := limiter.Take(req)
	if !res3.Allowed || res3.Remaining != 0 {
		t.Errorf("req 3: expected allowed=true, remaining=0; got allowed=%v, remaining=%d", res3.Allowed, res3.Remaining)
	}

	// Request 4: Blocked (Allowed = false)
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

	// Request 1: Status 200 OK
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
	if w1.Header().Get("X-RateLimit-Remaining") != "1" {
		t.Errorf("expected X-RateLimit-Remaining=1, got %s", w1.Header().Get("X-RateLimit-Remaining"))
	}

	// Request 2: Status 200 OK (Remaining = 0)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.1:5555"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w2.Code)
	}

	// Request 3: Status 429 Too Many Requests
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.RemoteAddr = "10.0.0.1:5555"
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", w3.Code)
	}
	if w3.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}
}

func TestKeyExtractors(t *testing.T) {
	// Case 1: KeyByIP with X-Forwarded-For
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18")
	if ip := ratelimit.KeyByIP(req1); ip != "203.0.113.195" {
		t.Errorf("expected '203.0.113.195', got '%s'", ip)
	}

	// Case 2: KeyByHeader
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-API-Key", "secret-key-123")
	keyFn := ratelimit.KeyByHeader("X-API-Key")
	if key := keyFn(req2); key != "X-API-Key:secret-key-123" {
		t.Errorf("expected 'X-API-Key:secret-key-123', got '%s'", key)
	}

	// Case 3: KeyByPathAndIP
	req3 := httptest.NewRequest(http.MethodPost, "/login", nil)
	req3.RemoteAddr = "192.168.1.1:8080"
	if key := ratelimit.KeyByPathAndIP(req3); key != "/login:192.168.1.1" {
		t.Errorf("expected '/login:192.168.1.1', got '%s'", key)
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

	// 50 * 20 = 1000 total requests. Next request should be blocked.
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

type dummyRedisClient struct{}

func (d *dummyRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	// Return mock Lua script response: [1, 5] (Allowed=1, Remaining=5)
	return []interface{}{int64(1), int64(5)}, nil
}

func TestRedisStore_Constructor(t *testing.T) {
	dummy := &dummyRedisClient{}
	redisStore := ratelimit.NewRedisStore(dummy, "test_prefix")

	limiter, err := ratelimit.New(redisStore, ratelimit.Rule{Capacity: 10, Window: 1 * time.Minute})
	if err != nil {
		t.Fatalf("failed to create redis limiter: %v", err)
	}
	defer limiter.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	res, err := limiter.Take(req)

	if err != nil {
		t.Fatalf("unexpected redis take error: %v", err)
	}
	if !res.Allowed || res.Remaining != 5 {
		t.Errorf("expected allowed=true, remaining=5; got allowed=%v, remaining=%d", res.Allowed, res.Remaining)
	}
}

func TestLimiter_CustomErrorHandler(t *testing.T) {
	var customCalled bool

	limiter, _ := ratelimit.New(
		ratelimit.NewMemoryStore(),
		ratelimit.Rule{Capacity: 1, Window: 1 * time.Minute},
		ratelimit.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, result ratelimit.Result) {
			customCalled = true
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"custom": true}`))
		}),
	)
	defer limiter.Close()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Req 1: Allowed
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	// Req 2: Blocked -> Custom Error Handler
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if !customCalled {
		t.Errorf("expected custom error handler to be called")
	}
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}
}

func BenchmarkMemoryStore_Take(b *testing.B) {
	limiter, _ := ratelimit.New(
		ratelimit.NewMemoryStore(),
		ratelimit.Rule{Capacity: 1000000, Window: 1 * time.Minute},
	)
	defer limiter.Close()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = limiter.Take(req)
	}
}
