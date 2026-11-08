package gometrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCounter_Concurrent(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup
	const routines = 100
	const perRoutine = 1000

	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perRoutine; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	expected := float64(routines * perRoutine)
	if c.Value() != expected {
		t.Fatalf("expected counter %f, got %f", expected, c.Value())
	}
}

func TestGauge(t *testing.T) {
	g := NewGauge()
	g.Set(10.5)
	if g.Value() != 10.5 {
		t.Fatalf("expected 10.5, got %f", g.Value())
	}

	g.Inc()
	if g.Value() != 11.5 {
		t.Fatalf("expected 11.5, got %f", g.Value())
	}

	g.Dec()
	if g.Value() != 10.5 {
		t.Fatalf("expected 10.5, got %f", g.Value())
	}

	g.Add(5.0)
	if g.Value() != 15.5 {
		t.Fatalf("expected 15.5, got %f", g.Value())
	}
}

func TestHistogram(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 5.0}
	h := NewHistogram(buckets)

	// Add observations
	h.Observe(0.05) // <= 0.1, 0.5, 1.0, 5.0
	h.Observe(0.25) // <= 0.5, 1.0, 5.0
	h.Observe(0.75) // <= 1.0, 5.0
	h.Observe(2.50) // <= 5.0

	count, sum, bList, counts := h.Snapshot()
	if count != 4 {
		t.Fatalf("expected count 4, got %d", count)
	}
	if sum != 3.55 {
		t.Fatalf("expected sum 3.55, got %f", sum)
	}
	if len(bList) != 4 || counts[0] != 1 || counts[1] != 2 || counts[2] != 3 || counts[3] != 4 {
		t.Fatalf("unexpected bucket counts: %+v", counts)
	}

	// Test approximate percentiles
	p50 := h.Percentile(50.0)
	if p50 <= 0 || p50 > 5.0 {
		t.Fatalf("percentile 50 unexpected: %f", p50)
	}
}

func TestRegistry_Gather(t *testing.T) {
	reg := NewRegistry()

	c := reg.Counter("orders_total", "Total completed orders.", Labels{"status": "paid"})
	c.Add(12)

	g := reg.Gauge("active_users", "Current online users.", nil)
	g.Set(42)

	h := reg.Histogram("payment_latency_seconds", "Payment API response latency.", []float64{0.1, 1.0}, Labels{"gateway": "razorpay"})
	h.Observe(0.05)
	h.Observe(0.50)

	text := reg.Gather()

	// Assert Prometheus text output format
	if !strings.Contains(text, "# HELP orders_total Total completed orders.") {
		t.Fatal("missing HELP header for orders_total")
	}
	if !strings.Contains(text, "# TYPE orders_total counter") {
		t.Fatal("missing TYPE header for orders_total")
	}
	if !strings.Contains(text, `orders_total{status="paid"} 12`) {
		t.Fatalf("missing orders_total metric line, got:\n%s", text)
	}
	if !strings.Contains(text, "active_users 42") {
		t.Fatalf("missing active_users metric line, got:\n%s", text)
	}
	if !strings.Contains(text, `payment_latency_seconds_bucket{gateway="razorpay",le="0.1"} 1`) {
		t.Fatalf("missing histogram bucket, got:\n%s", text)
	}
	if !strings.Contains(text, `payment_latency_seconds_count{gateway="razorpay"} 2`) {
		t.Fatalf("missing histogram count, got:\n%s", text)
	}
}

func TestHTTPHandlerAndMiddleware(t *testing.T) {
	reg := NewRegistry()
	sampler := RegisterRuntimeMetrics(reg)

	// Target handler wrapped in middleware
	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})

	mw := HTTPMiddleware(reg)(targetHandler)

	// Simulate incoming HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rec.Code)
	}

	// Now query /metrics endpoint
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	Handler(reg, sampler).ServeHTTP(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /metrics, got %d", metricsRec.Code)
	}

	body := metricsRec.Body.String()
	if !strings.Contains(body, `http_requests_total{method="POST",status="201"} 1`) {
		t.Fatalf("missing http_requests_total metric in /metrics:\n%s", body)
	}
	if !strings.Contains(body, "go_goroutines") {
		t.Fatalf("missing runtime metric in /metrics:\n%s", body)
	}
}
