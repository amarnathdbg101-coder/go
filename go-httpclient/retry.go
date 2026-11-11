package gohttpclient

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"net/http"
	"strconv"
	"time"
)

// RetryPolicy determines whether an HTTP request should be retried based on response or error.
type RetryPolicy func(resp *http.Response, err error) bool

// OnRetryFunc is a callback invoked before a retry sleep occurs.
type OnRetryFunc func(attempt int, req *http.Request, resp *http.Response, err error)

// RetryConfig defines exponential backoff with full jitter retry parameters.
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RetryableStatus map[int]bool
	RetryPolicy     RetryPolicy
}

// DefaultRetryConfig sets resilient production defaults.
var DefaultRetryConfig = RetryConfig{
	MaxRetries:      3,
	InitialInterval: 100 * time.Millisecond,
	MaxInterval:     2 * time.Second,
	Multiplier:      2.0,
	RetryableStatus: map[int]bool{
		http.StatusTooManyRequests:     true, // 429
		http.StatusInternalServerError: true, // 500
		http.StatusBadGateway:          true, // 502
		http.StatusServiceUnavailable:  true, // 503
		http.StatusGatewayTimeout:       true, // 504
	},
}

// BackoffDuration calculates exponential backoff duration with full jitter for attempt number (0-indexed).
func (rc *RetryConfig) BackoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		return rc.InitialInterval
	}

	mult := 1.0
	for i := 0; i < attempt; i++ {
		mult *= rc.Multiplier
	}

	calculated := float64(rc.InitialInterval) * mult
	if calculated > float64(rc.MaxInterval) {
		calculated = float64(rc.MaxInterval)
	}

	// Full Jitter with cryptographic random seed to avoid herd effects
	var b [8]byte
	_, _ = rand.Read(b[:])
	r := binary.LittleEndian.Uint64(b[:])
	jitter := time.Duration(r % uint64(calculated))

	if jitter < rc.InitialInterval {
		return rc.InitialInterval
	}
	return jitter
}

// CalculateBackoff computes backoff considering Retry-After header if present on the response.
func (rc *RetryConfig) CalculateBackoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
			// 1. Integer seconds format (e.g. "120")
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil && seconds > 0 {
				return time.Duration(seconds) * time.Second
			}
			// 2. HTTP-date format (e.g. "Fri, 31 Dec 2026 23:59:59 GMT")
			if targetTime, err := http.ParseTime(retryAfterStr); err == nil {
				if wait := time.Until(targetTime); wait > 0 {
					return wait
				}
			}
		}
	}
	return rc.BackoffDuration(attempt)
}

// CheckRetry evaluates whether an HTTP response or error qualifies for a retry.
func (rc *RetryConfig) CheckRetry(resp *http.Response, err error) bool {
	if rc.RetryPolicy != nil {
		return rc.RetryPolicy(resp, err)
	}
	var statusCode int
	if resp != nil {
		statusCode = resp.StatusCode
	}
	return rc.IsRetryable(statusCode, err)
}

// IsRetryable determines whether an error or status code qualifies for a retry using status map.
func (rc *RetryConfig) IsRetryable(statusCode int, err error) bool {
	if err != nil {
		// Network connection refused, reset, or timeout
		if netErr, ok := err.(net.Error); ok {
			return netErr.Timeout()
		}
		return true
	}

	return rc.RetryableStatus[statusCode]
}
