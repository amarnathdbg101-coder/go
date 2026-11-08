package ratelimit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// ErrorHandler is called when a request exceeds rate limit capacity.
// Allows customizing HTTP status code and response payload.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, result Result)

// DefaultErrorHandler renders a clean JSON HTTP 429 Too Many Requests response.
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, result Result) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if result.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
	}
	w.WriteHeader(http.StatusTooManyRequests)

	retrySec := int(result.RetryAfter.Seconds())
	if retrySec <= 0 {
		retrySec = 1
	}

	msg := fmt.Sprintf("Rate limit exceeded. Please retry in %d seconds.", retrySec)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   "Too Many Requests",
		"message": msg,
	})
}

// Middleware returns a standard net/http middleware HTTP handler wrapper.
// Works seamlessly with Chi, Gorilla, net/http, Gin, Fiber, or Echo.
func (e *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := e.Take(r)
		if err != nil {
			// On storage failure, fail open (allow request) to prevent blocking legitimate users
			next.ServeHTTP(w, r)
			return
		}

		// Inject standard RFC 6585 / IETF Rate Limit Headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(e.rule.Capacity))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(res.ResetAt.Unix(), 10))

		if !res.Allowed {
			if e.onLimitExceeded != nil {
				e.onLimitExceeded(w, r, res)
			} else {
				DefaultErrorHandler(w, r, res)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HandlerFunc wraps a standard http.HandlerFunc with rate limiting.
func (e *Limiter) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return e.Middleware(http.HandlerFunc(next)).ServeHTTP
}
