package gowebhook

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestSignAndVerify_Success(t *testing.T) {
	secret := []byte("super-secret-key-12345")
	payload := []byte(`{"event":"order.created","amount":1500}`)
	now := time.Now().Unix()

	sigHeader, err := Sign(secret, payload, now)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	err = Verify(secret, payload, sigHeader, 5*time.Minute)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
}

func TestVerify_TamperedPayload(t *testing.T) {
	secret := []byte("secret")
	originalPayload := []byte(`{"id":101}`)
	tamperedPayload := []byte(`{"id":102}`)

	sigHeader, _ := Sign(secret, originalPayload, time.Now().Unix())

	err := Verify(secret, tamperedPayload, sigHeader, 5*time.Minute)
	if err == nil {
		t.Errorf("expected verification failure on tampered payload, got nil")
	}
}

func TestVerify_ExpiredTimestamp(t *testing.T) {
	secret := []byte("secret")
	payload := []byte(`{"id":101}`)
	oldTimestamp := time.Now().Add(-10 * time.Minute).Unix()

	sigHeader, _ := Sign(secret, payload, oldTimestamp)

	// Tolerance is 1 minute, signature is 10 minutes old
	err := Verify(secret, payload, sigHeader, 1*time.Minute)
	if err != ErrSignatureExpired {
		t.Errorf("expected ErrSignatureExpired, got %v", err)
	}
}

func TestDispatcher_SynchronousDelivery(t *testing.T) {
	secret := []byte("test-secret-456")
	payload := []byte(`{"event":"user.signup"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sig := r.Header.Get("X-Webhook-Signature")
		event := r.Header.Get("X-Webhook-Event")
		id := r.Header.Get("X-Webhook-ID")

		if event != "user.signup" || id != "wh_1001" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := Verify(secret, body, sig, 1*time.Minute); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer server.Close()

	d := NewDispatcher(DefaultConfig())
	defer d.Stop()

	res, err := d.Dispatch(context.Background(), server.URL, "user.signup", "wh_1001", payload, secret)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if !res.Success || res.StatusCode != http.StatusOK {
		t.Errorf("expected success HTTP 200, got status %d", res.StatusCode)
	}
	if res.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", res.Attempts)
	}
}

func TestDispatcher_RetriesOnServerError(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"event":"test"}`)

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		att := attempts.Add(1)
		if att < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BackoffBase = 10 * time.Millisecond
	d := NewDispatcher(cfg)
	defer d.Stop()

	res, err := d.Dispatch(context.Background(), server.URL, "test", "wh_retry", payload, secret)
	if err != nil {
		t.Fatalf("Dispatch retry failed: %v", err)
	}

	if !res.Success || res.Attempts != 3 {
		t.Errorf("expected success after 3 attempts, got attempts=%d, success=%v", res.Attempts, res.Success)
	}
}

func TestDispatcher_AsyncDispatch(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"async":true}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	d := NewDispatcher(DefaultConfig())
	defer d.Stop()

	done := make(chan bool)
	d.DispatchAsync(context.Background(), server.URL, "async.event", "wh_async", payload, secret, func(res *DeliveryResult) {
		if res.Success {
			done <- true
		} else {
			done <- false
		}
	})

	select {
	case success := <-done:
		if !success {
			t.Errorf("async dispatch reported failure")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("async dispatch timed out")
	}
}
