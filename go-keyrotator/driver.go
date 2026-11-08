package keyrotator

import "time"

// KeyHealth holds runtime observability metrics and health status for an individual credential key.
type KeyHealth struct {
	Key               string    `json:"key"`
	ConsecutiveErrors int       `json:"consecutive_errors"`
	IsRateLimited     bool      `json:"is_rate_limited"`
	RateLimitedUntil  time.Time `json:"rate_limited_until"`
	TotalCalls        int64     `json:"total_calls"`
	TotalSuccess      int64     `json:"total_success"`
	TotalErrors       int64     `json:"total_errors"`
	LastUsedAt        time.Time `json:"last_used_at"`
}

// Strategy defines key selection algorithms (e.g. RoundRobin, LeastUsed).
type Strategy interface {
	// Select chooses the next viable key from the candidate list.
	Select(keys []*KeyHealth) (*KeyHealth, error)
}
