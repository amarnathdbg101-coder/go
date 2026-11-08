package gojwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var defaultHeaderB64 = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

// Engine signs and verifies HMAC-SHA256 (HS256) JSON Web Tokens.
type Engine struct {
	secret []byte
	leeway time.Duration
}

// New constructs an Engine with secret key and optional clock skew tolerance.
func New(secret []byte, leeway ...time.Duration) (*Engine, error) {
	if len(secret) == 0 {
		return nil, ErrMissingSecret
	}

	var l time.Duration
	if len(leeway) > 0 {
		l = leeway[0]
	}

	secCopy := make([]byte, len(secret))
	copy(secCopy, secret)

	return &Engine{
		secret: secCopy,
		leeway: l,
	}, nil
}

// Sign serializes and cryptographically signs a Claims struct into a JWT string.
func (e *Engine) Sign(claims Claims) (string, error) {
	if claims.IssuedAt == 0 {
		claims.IssuedAt = time.Now().Unix()
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("jwt: marshal claims: %w", err)
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := defaultHeaderB64 + "." + payloadB64

	h := hmac.New(sha256.New, e.secret)
	h.Write([]byte(signingInput))
	signature := h.Sum(nil)
	sigB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + sigB64, nil
}

// Verify decodes, authenticates, and validates a JWT token string.
func (e *Engine) Verify(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerB64, payloadB64, sigB64 := parts[0], parts[1], parts[2]
	signingInput := headerB64 + "." + payloadB64

	// Verify HMAC-SHA256 signature in constant time
	expectedSigBytes := hmacSHA256([]byte(signingInput), e.secret)
	actualSigBytes, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return nil, ErrInvalidSignature
	}

	if subtle.ConstantTimeCompare(expectedSigBytes, actualSigBytes) != 1 {
		return nil, ErrInvalidSignature
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("jwt: unmarshal claims: %w", err)
	}

	// Validate temporal validity
	if err := claims.Valid(e.leeway); err != nil {
		return nil, err
	}

	return &claims, nil
}

func hmacSHA256(message, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(message)
	return h.Sum(nil)
}

// Sign creates a signed JWT using Default settings.
func Sign(secret []byte, claims Claims) (string, error) {
	e, err := New(secret)
	if err != nil {
		return "", err
	}
	return e.Sign(claims)
}

// Verify authenticates a JWT using Default settings.
func Verify(secret []byte, tokenStr string) (*Claims, error) {
	e, err := New(secret)
	if err != nil {
		return nil, err
	}
	return e.Verify(tokenStr)
}
