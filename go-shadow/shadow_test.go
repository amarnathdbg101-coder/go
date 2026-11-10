package shadow_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-shadow"
)

func TestShadowDiffDetection(t *testing.T) {
	// 1. Mock candidate server returning different JSON
	candidate := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version":"v2","count":10}`))
	}))
	defer candidate.Close()

	var diffCount int32
	var capturedDiff *shadow.DiffReport
	var mu sync.Mutex

	engine := shadow.New(
		shadow.WithTargetURL(candidate.URL),
		shadow.WithSampleRate(1.0),
		shadow.WithOnDiff(func(diff *shadow.DiffReport) {
			atomic.AddInt32(&diffCount, 1)
			mu.Lock()
			capturedDiff = diff
			mu.Unlock()
		}),
	)
	defer engine.Close()

	// 2. Primary handler returning v1 JSON
	primaryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version":"v1","count":10}`))
	})

	middleware := shadow.Middleware(engine)
	handler := middleware(primaryHandler)

	req := httptest.NewRequest("GET", "/api/data", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// User received primary response
	if !strings.Contains(rr.Body.String(), `"version":"v1"`) {
		t.Fatalf("primary user should receive v1 response, got: %s", rr.Body.String())
	}

	// Wait for background shadow comparison
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&diffCount) != 1 {
		t.Fatalf("expected 1 diff report triggered, got %d", atomic.LoadInt32(&diffCount))
	}

	mu.Lock()
	defer mu.Unlock()
	if !capturedDiff.BodyMismatch {
		t.Errorf("expected BodyMismatch to be true")
	}
}

func TestStatusMismatch(t *testing.T) {
	candidate := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer candidate.Close()

	var diffTriggered int32
	engine := shadow.New(
		shadow.WithTargetURL(candidate.URL),
		shadow.WithOnDiff(func(diff *shadow.DiffReport) {
			if diff.StatusMismatch {
				atomic.AddInt32(&diffTriggered, 1)
			}
		}),
	)
	defer engine.Close()

	primary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler := shadow.Middleware(engine)(primary)

	req := httptest.NewRequest("POST", "/checkout", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&diffTriggered) != 1 {
		t.Fatalf("expected status mismatch to be triggered")
	}
}

