package gocookie

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Option tunes Cookie Jar configuration.
type Option func(*Jar)

// WithInsecureForDev allows HTTP (non-secure) cookies for local development.
func WithInsecureForDev() Option {
	return func(j *Jar) {
		j.secure = false
	}
}

// WithSameSite configures the default SameSite policy.
func WithSameSite(mode http.SameSite) Option {
	return func(j *Jar) {
		j.sameSite = mode
	}
}

// Jar securely encrypts, signs, and reads HTTP cookies using AES-256-GCM.
type Jar struct {
	gcm      cipher.AEAD
	secure   bool
	sameSite http.SameSite
}

// New constructs a secure Jar requiring a 32-byte (256-bit) encryption key.
func New(secretKey32 []byte, opts ...Option) (*Jar, error) {
	if len(secretKey32) != 32 {
		return nil, fmt.Errorf("%w: got %d bytes", ErrInvalidKeyLength, len(secretKey32))
	}

	block, err := aes.NewCipher(secretKey32)
	if err != nil {
		return nil, fmt.Errorf("cookie: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cookie: new gcm: %w", err)
	}

	j := &Jar{
		gcm:      gcm,
		secure:   true, // Security by default
		sameSite: http.SameSiteLaxMode,
	}

	for _, opt := range opts {
		opt(j)
	}

	return j, nil
}

// Write encrypts cookie.Value with AES-256-GCM, applies security headers (HttpOnly, Secure, SameSite), and writes to ResponseWriter.
func (j *Jar) Write(w http.ResponseWriter, cookie http.Cookie) error {
	nonce := make([]byte, j.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("cookie: generate nonce: %w", err)
	}

	// Seal plaintext: [nonce | ciphertext + tag]
	ciphertext := j.gcm.Seal(nonce, nonce, []byte(cookie.Value), []byte(cookie.Name))
	encoded := base64.RawURLEncoding.EncodeToString(ciphertext)

	cookie.Value = encoded
	cookie.HttpOnly = true
	cookie.Secure = j.secure
	cookie.SameSite = j.sameSite
	if cookie.Path == "" {
		cookie.Path = "/"
	}

	http.SetCookie(w, &cookie)
	return nil
}

// Read extracts and decrypts an authenticated cookie by name from the HTTP request.
func (j *Jar) Read(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrCookieNotFound
		}
		return "", err
	}

	data, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCookieValue, err)
	}

	nonceSize := j.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrTamperedCookie
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := j.gcm.Open(nil, nonce, ciphertext, []byte(cookie.Name))
	if err != nil {
		return "", ErrTamperedCookie
	}

	return string(plaintext), nil
}

// Delete invalidates and expires a cookie in the client browser.
func (j *Jar) Delete(w http.ResponseWriter, name string) {
	cookie := http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   j.secure,
		SameSite: j.sameSite,
	}
	http.SetCookie(w, &cookie)
}
