package secguard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecGuard_HeadersInjected(t *testing.T) {
	eng := New()
	defer eng.Close()

	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing nosniff header")
	}
	if rec.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("missing X-Frame-Options header")
	}
}

func TestSecGuard_SQLiBlockAndJail(t *testing.T) {
	eng := New(WithJailDuration(5 * time.Minute))
	defer eng.Close()

	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Attempt SQL injection via query parameter
	req := httptest.NewRequest(http.MethodGet, "/search?q=1+union+select+*+from+users", nil)
	req.RemoteAddr = "192.168.1.105:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for SQLi attempt, got %d", rec.Code)
	}

	// Subsequent legitimate request from the same IP must now be blocked due to quarantine jail
	req2 := httptest.NewRequest(http.MethodGet, "/normal", nil)
	req2.RemoteAddr = "192.168.1.105:54321"
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for quarantined IP, got %d", rec2.Code)
	}
	if rec2.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header on quarantined response")
	}
}

func TestSecGuard_ExploitProbeBlock(t *testing.T) {
	eng := New()
	defer eng.Close()

	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	req.RemoteAddr = "10.0.0.50:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for /.env probe, got %d", rec.Code)
	}
}

func TestSecGuard_ScannerBlock(t *testing.T) {
	eng := New()
	defer eng.Close()

	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "sqlmap/1.4.10#stable")
	req.RemoteAddr = "10.0.0.88:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for sqlmap scanner, got %d", rec.Code)
	}
}

func TestSecGuard_ExemptPath(t *testing.T) {
	eng := New()
	defer eng.Close()

	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthy"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for health endpoint, got %d", rec.Code)
	}
}

func TestSecGuard_Honeypot(t *testing.T) {
	jail := NewMemoryJail(100, 1*time.Minute)
	defer jail.Close()

	// Legitimate human (honeypot empty)
	if !ValidateHoneypot("", "192.168.1.1", jail, 10*time.Minute) {
		t.Fatal("expected legitimate user to pass honeypot")
	}

	// Spambot (honeypot populated)
	if ValidateHoneypot("buy cheap meds", "192.168.1.200", jail, 10*time.Minute) {
		t.Fatal("expected bot to fail honeypot")
	}

	jailed, _, _, _ := jail.IsJailed(context.Background(), "192.168.1.200")
	if !jailed {
		t.Fatal("expected bot IP to be jailed")
	}
}

func TestSecGuard_ObservabilityHooks(t *testing.T) {
	var capturedWAFRule string
	var capturedScanner string
	var capturedProbe string
	var jailedIP string

	eng := New(WithHooks(Hooks{
		OnWAFViolation: func(ip, rule, path string) {
			capturedWAFRule = rule
		},
		OnScannerBlocked: func(ip, ua string) {
			capturedScanner = ua
		},
		OnExploitProbe: func(ip, path string) {
			capturedProbe = path
		},
		OnIPJailed: func(ip string, d time.Duration, reason string) {
			jailedIP = ip
		},
	}))
	defer eng.Close()

	handler := eng.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Trigger WAF hook
	req1 := httptest.NewRequest(http.MethodGet, "/search?q=1+union+select+*", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if capturedWAFRule == "" || jailedIP != "10.0.0.1" {
		t.Fatal("OnWAFViolation or OnIPJailed hook failed")
	}

	// 2. Trigger Scanner hook
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("User-Agent", "nikto/2.1.6")
	req2.RemoteAddr = "10.0.0.2:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if capturedScanner != "nikto/2.1.6" {
		t.Fatal("OnScannerBlocked hook failed")
	}

	// 3. Trigger Exploit Probe hook
	req3 := httptest.NewRequest(http.MethodGet, "/.env", nil)
	req3.RemoteAddr = "10.0.0.3:1234"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	if capturedProbe != "/.env" {
		t.Fatal("OnExploitProbe hook failed")
	}
}
