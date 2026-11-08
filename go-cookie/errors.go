package gocookie

import "errors"

var (
	// ErrInvalidKeyLength occurs when encryption key is not 32 bytes (for AES-256).
	ErrInvalidKeyLength = errors.New("cookie: secret key must be exactly 32 bytes")

	// ErrCookieNotFound occurs when requested cookie is missing from the HTTP request.
	ErrCookieNotFound = errors.New("cookie: cookie not found in request")

	// ErrInvalidCookieValue occurs when cookie value is malformed or not Base64 encoded.
	ErrInvalidCookieValue = errors.New("cookie: invalid base64 cookie payload")

	// ErrTamperedCookie occurs when AES-GCM tag verification fails (data tampered).
	ErrTamperedCookie = errors.New("cookie: authentication failed, cookie tampered or corrupt")
)
