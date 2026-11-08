package gocontextkit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

type contextKey int

const (
	keyRequestID contextKey = iota
	keyCorrelationID
	keyUserID
	keyTenantID
	keyClientIP
	keyMetadata
)

// WithRequestID returns a new context with the request ID attached.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

// GetRequestID retrieves the request ID from ctx or empty string.
func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(keyRequestID).(string); ok {
		return val
	}
	return ""
}

// WithCorrelationID returns a new context with the distributed correlation ID attached.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyCorrelationID, id)
}

// GetCorrelationID retrieves the correlation ID from ctx or empty string.
func GetCorrelationID(ctx context.Context) string {
	if val, ok := ctx.Value(keyCorrelationID).(string); ok {
		return val
	}
	return ""
}

// WithUserID returns a new context with the authenticated user ID.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, keyUserID, userID)
}

// GetUserID retrieves the user ID from ctx or empty string.
func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(keyUserID).(string); ok {
		return val
	}
	return ""
}

// WithTenantID returns a new context with multi-tenant workspace/organization ID.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, keyTenantID, tenantID)
}

// GetTenantID retrieves the tenant ID from ctx or empty string.
func GetTenantID(ctx context.Context) string {
	if val, ok := ctx.Value(keyTenantID).(string); ok {
		return val
	}
	return ""
}

// WithClientIP returns a new context with the client IP address.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, keyClientIP, ip)
}

// GetClientIP retrieves the client IP from ctx or empty string.
func GetClientIP(ctx context.Context) string {
	if val, ok := ctx.Value(keyClientIP).(string); ok {
		return val
	}
	return ""
}

// WithMetadata attaches arbitrary key-value metadata to the context.
func WithMetadata(ctx context.Context, key, val string) context.Context {
	m, ok := ctx.Value(keyMetadata).(map[string]string)
	if !ok || m == nil {
		m = make(map[string]string)
	}
	newMap := make(map[string]string, len(m)+1)
	for k, v := range m {
		newMap[k] = v
	}
	newMap[key] = val
	return context.WithValue(ctx, keyMetadata, newMap)
}

// GetMetadata retrieves a key-value pair from context metadata.
func GetMetadata(ctx context.Context, key string) (string, bool) {
	if m, ok := ctx.Value(keyMetadata).(map[string]string); ok && m != nil {
		val, exists := m[key]
		return val, exists
	}
	return "", false
}

type detachedContext struct {
	parent context.Context
}

// Detach creates a context that retains all values of parent but ignores parent cancellation or deadlines.
// Essential for spawning background goroutines from HTTP handlers without premature termination upon HTTP response finish.
func Detach(ctx context.Context) context.Context {
	return &detachedContext{parent: ctx}
}

func (d *detachedContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (d *detachedContext) Done() <-chan struct{} {
	return nil
}

func (d *detachedContext) Err() error {
	return nil
}

func (d *detachedContext) Value(key any) any {
	return d.parent.Value(key)
}

// GenerateID produces a secure 16-character hex ID for request tracing.
func GenerateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// HTTPMiddleware extracts or assigns X-Request-ID, captures client IP, and injects both into ctx and response headers.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = GenerateID()
		}

		clientIP := extractClientIP(r)

		ctx := r.Context()
		ctx = WithRequestID(ctx, reqID)
		ctx = WithClientIP(ctx, clientIP)

		// Mirror to response header for distributed debugging
		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	// Fallback to RemoteAddr (strip port if present)
	addr := r.RemoteAddr
	if colon := strings.LastIndex(addr, ":"); colon != -1 {
		return addr[:colon]
	}
	return addr
}
