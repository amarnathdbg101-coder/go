package ratelimit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request, result Result)

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

func (e *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := e.Take(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

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

func (e *Limiter) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return e.Middleware(http.HandlerFunc(next)).ServeHTTP
}
