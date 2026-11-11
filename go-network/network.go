package network

import (
	"time"

	"github.com/shopsilo/go-network/ratelimit"
)

// Version of the unified go-network suite.
const Version = "1.0.0"

// Re-exported types for top-level convenience.
type RateLimiter = ratelimit.Limiter
type RateRule = ratelimit.Rule

// NewRateLimiter initializes an IP-based rate limiter.
func NewRateLimiter(capacity int, windowSeconds int) (*ratelimit.Limiter, error) {
	return ratelimit.NewIPLimiter(capacity, time.Duration(windowSeconds)*time.Second)
}
