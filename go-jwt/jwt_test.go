package gojwt

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestJWT_SignAndVerify(t *testing.T) {
	secret := []byte("production_super_secret_key_2026")
	engine, err := New(secret)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	claims := Claims{
		Subject:   "usr_102",
		Issuer:    "ShopSilo-Auth",
		Audience:  "ShopSilo-Web",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		Custom: map[string]any{
			"role":  "admin",
			"email": "aman@example.com",
		},
	}

	tokenStr, err := engine.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Must be 3 parts
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts in JWT, got %d", len(parts))
	}

	// Verify
	parsed, err := engine.Verify(tokenStr)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if parsed.Subject != "usr_102" || parsed.Issuer != "ShopSilo-Auth" {
		t.Fatalf("claims mismatch: %+v", parsed)
	}
	if parsed.Custom["role"] != "admin" {
		t.Fatalf("custom claim mismatch: %+v", parsed.Custom)
	}
}

func TestJWT_TamperedSignature(t *testing.T) {
	secret := []byte("secret_1")
	engine, _ := New(secret)

	token, _ := engine.Sign(Claims{Subject: "1", ExpiresAt: time.Now().Add(time.Hour).Unix()})

	// Guaranteed deterministic alteration of the signature
	lastChar := token[len(token)-1]
	replacement := "a"
	if lastChar == 'a' {
		replacement = "b"
	}
	tampered := token[:len(token)-1] + replacement

	_, err := engine.Verify(tampered)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	engine1, _ := New([]byte("secret_correct"))
	engine2, _ := New([]byte("secret_wrong"))

	token, _ := engine1.Sign(Claims{Subject: "1", ExpiresAt: time.Now().Add(time.Hour).Unix()})

	_, err := engine2.Verify(token)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature on wrong secret, got %v", err)
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	secret := []byte("secret")
	engine, _ := New(secret)

	// Expired token (1 hour ago)
	token, _ := engine.Sign(Claims{
		Subject:   "1",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	})

	_, err := engine.Verify(token)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}

	// Token expired 2 seconds ago, with 2-minute leeway, it should be accepted
	recentExpiredToken, _ := engine.Sign(Claims{
		Subject:   "1",
		ExpiresAt: time.Now().Add(-2 * time.Second).Unix(),
	})

	engineWithLeeway, _ := New(secret, 2*time.Minute)
	parsed, err := engineWithLeeway.Verify(recentExpiredToken)
	if err != nil || parsed == nil {
		t.Fatalf("expected token to be valid within leeway, err: %v", err)
	}
}

func TestJWT_NotBefore(t *testing.T) {
	secret := []byte("secret")
	engine, _ := New(secret)

	// Valid 1 hour in future
	token, _ := engine.Sign(Claims{
		Subject:   "1",
		NotBefore: time.Now().Add(1 * time.Hour).Unix(),
		ExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
	})

	_, err := engine.Verify(token)
	if !errors.Is(err, ErrTokenNotValidYet) {
		t.Fatalf("expected ErrTokenNotValidYet, got %v", err)
	}
}
