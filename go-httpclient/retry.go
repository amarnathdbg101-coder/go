package gohttpclient

import (
	"math/rand"
	"net"
	"net/http"
	"time"
)

// RetryConfig defines exponential backoff with full jitter retry parameters.
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RetryableStatus map[int]bool
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

	// Full Jitter: random duration between 0 and calculated
	jitter := rand.Float64() * calculated
	return time.Duration(jitter)
}

// IsRetryable determines whether an error or status code qualifies for a retry.
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
