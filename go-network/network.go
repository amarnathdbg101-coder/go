package network

import (
	"time"

	"github.com/amarnathdbg101-coder/go/go-network/ratelimit"
)

const Version = "1.0.0"

type RateLimiter = ratelimit.Limiter
type RateRule = ratelimit.Rule

func NewRateLimiter(capacity int, windowSeconds int) (*ratelimit.Limiter, error) {
	return ratelimit.NewIPLimiter(capacity, time.Duration(windowSeconds)*time.Second)
}
