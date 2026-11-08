package secguard

import (
	"context"
	"time"
)

// JailEntry tracks a quarantined client IP.
type JailEntry struct {
	IP        string    `json:"ip"`
	Reason    string    `json:"reason"`
	ExpiresAt time.Time `json:"expires_at"`
}

// JailStorage defines interface contracts for storing quarantined IPs.
type JailStorage interface {
	// Jail places an offending IP into quarantine for the specified duration.
	Jail(ctx context.Context, ip string, duration time.Duration, reason string) error

	// IsJailed checks whether an IP is currently banned and returns remaining seconds.
	IsJailed(ctx context.Context, ip string) (jailed bool, reason string, remainingSeconds int64, err error)

	// Unjail removes an IP from quarantine.
	Unjail(ctx context.Context, ip string) error

	// Close releases resources.
	Close() error
}
