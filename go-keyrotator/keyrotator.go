// Package keyrotator provides thread-safe credential pooling, round-robin rotation,
// and health tracking for AI API keys (Gemini, OpenAI), payment gateways, and proxies.
package keyrotator

import "strings"

// FromCSV creates a Rotator from a comma-separated string of keys.
func FromCSV(csv string, opts ...Option) *Rotator {
	keys := ParseList(csv)
	return New(keys, opts...)
}

// ParseList cleans and parses a comma-separated list of keys, stripping quotes and whitespace.
func ParseList(raw string) []string {
	if raw == "" {
		return []string{}
	}

	parts := strings.Split(raw, ",")
	res := make([]string, 0, len(parts))

	for _, p := range parts {
		clean := strings.TrimSpace(strings.Trim(p, `"'`))
		if clean != "" {
			res = append(res, clean)
		}
	}
	return res
}
