package gocookie

import (
	"crypto/rand"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookie_WriteAndRead(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	jar, err := New(key)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	rawSessionID := "sess_xyz_99482710492"

	// 1. Write Cookie
	rec := httptest.NewRecorder()
	err = jar.Write(rec, http.Cookie{
		Name:  "session_token",
		Value: rawSessionID,
	})
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	cookieHeader := rec.Header().Get("Set-Cookie")
	if cookieHeader == "" {
		t.Fatal("Set-Cookie header was not written")
	}

	// 2. Read back into simulated incoming request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Cookie", cookieHeader)

	val, err := jar.Read(req, "session_token")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if val != rawSessionID {
		t.Fatalf("expected %s, got %s", rawSessionID, val)
	}
}

func TestCookie_TamperingFails(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	jar, _ := New(key)

	rec := httptest.NewRecorder()
	_ = jar.Write(rec, http.Cookie{
		Name:  "role",
		Value: "user",
	})

	rawCookie := rec.Result().Cookies()[0]
	// Tamper value
	tamperedBytes := []byte(rawCookie.Value)
	tamperedBytes[len(tamperedBytes)-1] ^= 0xFF
	rawCookie.Value = string(tamperedBytes)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(rawCookie)

	_, err := jar.Read(req, "role")
	if !errors.Is(err, ErrTamperedCookie) {
		t.Fatalf("expected ErrTamperedCookie on altered payload, got %v", err)
	}
}

func TestCookie_NameSpoofingFails(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	jar, _ := New(key)

	rec := httptest.NewRecorder()
	_ = jar.Write(rec, http.Cookie{
		Name:  "admin_flag",
		Value: "true",
	})

	cookie := rec.Result().Cookies()[0]

	// Attacker copies encrypted payload into different cookie name
	spoofedCookie := &http.Cookie{
		Name:  "user_flag",
		Value: cookie.Value,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(spoofedCookie)

	_, err := jar.Read(req, "user_flag")
	if !errors.Is(err, ErrTamperedCookie) {
		t.Fatalf("expected ErrTamperedCookie on renamed cookie payload, got %v", err)
	}
}

func TestCookie_Delete(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	jar, _ := New(key)

	rec := httptest.NewRecorder()
	jar.Delete(rec, "session_token")

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie in response, got %d", len(cookies))
	}
	if cookies[0].MaxAge != -1 {
		t.Fatalf("expected MaxAge: -1, got %d", cookies[0].MaxAge)
	}
}

func TestCookie_InvalidKey(t *testing.T) {
	_, err := New([]byte("short-key"))
	if !errors.Is(err, ErrInvalidKeyLength) {
		t.Fatalf("expected ErrInvalidKeyLength, got %v", err)
	}
}
