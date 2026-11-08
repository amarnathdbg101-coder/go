package gohttpclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type UserPayload struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestClient_Get_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserPayload{
			ID:    42,
			Name:  "Aditya Sharma",
			Email: "aditya@example.com",
		})
	}))
	defer ts.Close()

	client := New(WithTimeout(2 * time.Second))

	var user UserPayload
	resp, err := client.Get(context.Background(), ts.URL+"/users/42", &user)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	if user.ID != 42 || user.Name != "Aditya Sharma" {
		t.Fatalf("unexpected user payload: %+v", user)
	}
}

func TestClient_Post_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-App-ID") != "ShopSilo-Core" {
			http.Error(w, "missing header", http.StatusForbidden)
			return
		}

		var payload UserPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		payload.ID = 101 // Server assigns ID
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	client := New(
		WithTimeout(2*time.Second),
		WithHeader("X-App-ID", "ShopSilo-Core"),
	)

	reqBody := UserPayload{Name: "Pooja Verma", Email: "pooja@example.com"}
	var created UserPayload

	resp, err := client.Post(context.Background(), ts.URL+"/users", reqBody, &created)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if created.ID != 101 || created.Name != "Pooja Verma" {
		t.Fatalf("unexpected response: %+v", created)
	}
}

func TestClient_Retry_SucceedsAfterFlapping(t *testing.T) {
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			// Fail first two attempts with 503
			http.Error(w, "server flapping", http.StatusServiceUnavailable)
			return
		}
		// 3rd attempt succeeds
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "recovered"})
	}))
	defer ts.Close()

	retryCfg := DefaultRetryConfig
	retryCfg.InitialInterval = 10 * time.Millisecond
	retryCfg.MaxInterval = 50 * time.Millisecond

	client := New(
		WithTimeout(2*time.Second),
		WithRetry(retryCfg),
	)

	var result map[string]string
	resp, err := client.Get(context.Background(), ts.URL+"/flapping", &result)
	if err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
	defer resp.Body.Close()

	if result["status"] != "recovered" {
		t.Fatalf("expected recovered status, got %v", result)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestClient_CircuitBreaker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fatal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	retryCfg := DefaultRetryConfig
	retryCfg.MaxRetries = 0 // No retry for fast circuit test

	client := New(
		WithTimeout(1*time.Second),
		WithRetry(retryCfg),
		WithCircuitBreaker(2, 500*time.Millisecond), // Trips after 2 failures
	)

	// Call 1 -> 500
	_, _ = client.Get(context.Background(), ts.URL, nil)
	// Call 2 -> 500 (Trips breaker)
	_, _ = client.Get(context.Background(), ts.URL, nil)

	// Call 3 -> Must fail immediately with ErrCircuitOpen WITHOUT touching server
	_, err := client.Get(context.Background(), ts.URL, nil)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestAddQueryParams(t *testing.T) {
	baseURL := "https://api.shopsilo.com/v1/orders"
	params := map[string]string{
		"page":   "2",
		"status": "pending",
	}

	fullURL, err := AddQueryParams(baseURL, params)
	if err != nil {
		t.Fatalf("AddQueryParams failed: %v", err)
	}

	if fullURL != "https://api.shopsilo.com/v1/orders?page=2&status=pending" &&
		fullURL != "https://api.shopsilo.com/v1/orders?status=pending&page=2" {
		t.Fatalf("unexpected full URL: %s", fullURL)
	}
}
