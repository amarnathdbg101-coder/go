// Package secguard provides zero-dependency OWASP HTTP security headers,
// in-memory WAF pattern inspection (SQLi, Traversal, XSS), automated scanner blocking,
// and auto-expiring IP quarantine jail.
package secguard

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// New creates an initialized SecGuard engine.
func New(opts ...Option) *Engine {
	return NewEngine(opts...)
}

// Middleware returns a standard net/http middleware handler.
func Middleware(opts ...Option) func(http.Handler) http.Handler {
	eng := NewEngine(opts...)
	return eng.Middleware
}

// ValidateHoneypot checks if a hidden honeypot form field was populated by a bot.
// If non-empty, automatically quarantines the IP and returns false.
func ValidateHoneypot(honeypotVal string, ip string, jail JailStorage, duration time.Duration) bool {
	if strings.TrimSpace(honeypotVal) != "" {
		if duration <= 0 {
			duration = 30 * time.Minute
		}
		if jail != nil {
			_ = jail.Jail(context.Background(), ip, duration, "honeypot trap triggered")
		}
		return false
	}
	return true
}
