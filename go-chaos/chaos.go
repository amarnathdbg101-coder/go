// Package chaos provides in-process chaos engineering and fault injection for Go HTTP servers,
// microservice clients, and resilience pipelines.
package chaos

import "time"

// Version of the go-chaos package.
const Version = "1.0.0"

// LatencyRule is a constructor for creating latency injection rules.
func LatencyRule(pathPrefix string, delay time.Duration, probability float64) *Rule {
	return &Rule{
		Type:        FaultLatency,
		Enabled:     true,
		PathPrefix:  pathPrefix,
		Delay:       delay,
		Probability: probability,
	}
}

// ErrorRule is a constructor for creating synthetic HTTP error rules.
func ErrorRule(pathPrefix string, statusCode int, probability float64) *Rule {
	return &Rule{
		Type:        FaultError,
		Enabled:     true,
		PathPrefix:  pathPrefix,
		StatusCode:  statusCode,
		Probability: probability,
	}
}
