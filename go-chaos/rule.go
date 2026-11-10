package chaos

import (
	"crypto/rand"
	"encoding/binary"
	"net/http"
	"strings"
	"time"
)

// FaultType specifies the category of failure to inject.
type FaultType string

const (
	FaultLatency FaultType = "LATENCY"
	FaultError   FaultType = "ERROR"
	FaultPanic   FaultType = "PANIC"
	FaultCorrupt FaultType = "CORRUPT"
)

// Rule defines a single fault injection behavior and its targeting criteria.
type Rule struct {
	ID          string        `json:"id"`
	Type        FaultType     `json:"type"`
	Enabled     bool          `json:"enabled"`
	Probability float64       `json:"probability"` // 0.0 to 1.0 (e.g. 0.25 = 25% chance)

	// Fault parameters
	Delay      time.Duration `json:"delay,omitempty"`
	StatusCode int           `json:"status_code,omitempty"`
	StatusBody string        `json:"status_body,omitempty"`
	PanicMsg   string        `json:"panic_msg,omitempty"`

	// Targeting criteria
	PathPrefix string `json:"path_prefix,omitempty"`
	HeaderKey  string `json:"header_key,omitempty"`
	HeaderVal  string `json:"header_val,omitempty"`
}

// Matches checks if an incoming HTTP request satisfies the targeting criteria and probability roll.
func (r *Rule) Matches(req *http.Request) bool {
	if !r.Enabled {
		return false
	}

	// 1. Path prefix check
	if r.PathPrefix != "" && !strings.HasPrefix(req.URL.Path, r.PathPrefix) {
		return false
	}

	// 2. Header match check
	if r.HeaderKey != "" {
		val := req.Header.Get(r.HeaderKey)
		if r.HeaderVal != "" && val != r.HeaderVal {
			return false
		}
		if r.HeaderVal == "" && val == "" {
			return false
		}
	}

	// 3. Probability roll
	if r.Probability <= 0 {
		return false
	}
	if r.Probability >= 1.0 {
		return true
	}

	return randomFloat() < r.Probability
}

// randomFloat returns a cryptographically secure random float in [0.0, 1.0).
func randomFloat() float64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	val := binary.BigEndian.Uint64(b[:])
	return float64(val) / float64(1<<64)
}
