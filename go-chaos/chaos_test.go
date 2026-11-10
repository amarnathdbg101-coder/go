package chaos_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-chaos"
)

func TestLatencyInjection(t *testing.T) {
	engine := chaos.New()
	engine.AddRule(chaos.LatencyRule("/api/v1/slow", 25*time.Millisecond, 1.0))

	handler := chaos.Middleware(engine)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/slow/items", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	handler.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	if elapsed < 20*time.Millisecond {
		t.Fatalf("expected injected delay >= 20ms, got %v", elapsed)
	}
	if rr.Header().Get("X-Chaos-Injected") != "LATENCY" {
		t.Errorf("expected X-Chaos-Injected header 'LATENCY'")
	}
}

func TestErrorInjection(t *testing.T) {
	engine := chaos.New()
	engine.AddRule(chaos.ErrorRule("/api/v1/payment", http.StatusServiceUnavailable, 1.0))

	handler := chaos.Middleware(engine)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("NORMAL_FLOW"))
	}))

	req := httptest.NewRequest("POST", "/api/v1/payment/checkout", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rr.Code)
	}

	body, _ := io.ReadAll(rr.Body)
	if string(body) == "NORMAL_FLOW" {
		t.Fatalf("handler should not have executed on error injection")
	}
}

func TestBypassHeader(t *testing.T) {
	engine := chaos.New(chaos.WithBypassHeader("X-Bypass-Chaos", "secret123"))
	engine.AddRule(chaos.ErrorRule("/api", 500, 1.0))

	handler := chaos.Middleware(engine)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("BYPASSED"))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Bypass-Chaos", "secret123")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected bypass status 200, got %d", rr.Code)
	}
}

func TestPanicInjection(t *testing.T) {
	engine := chaos.New()
	engine.AddRule(&chaos.Rule{
		Type:        chaos.FaultPanic,
		Enabled:     true,
		PathPrefix:  "/panic",
		Probability: 1.0,
		PanicMsg:    "synthetic panic test",
	})

	handler := chaos.Middleware(engine)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/panic/crash", nil)
	rr := httptest.NewRecorder()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from chaos engine, but none occurred")
		}
		if r != "synthetic panic test" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	handler.ServeHTTP(rr, req)
}

func TestMasterKillswitch(t *testing.T) {
	engine := chaos.New()
	engine.AddRule(chaos.ErrorRule("/api", 500, 1.0))

	// Disable master switch
	engine.SetEnabled(false)

	handler := chaos.Middleware(engine)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/data", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with master switch off, got %d", rr.Code)
	}
}

