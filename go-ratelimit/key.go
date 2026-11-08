package ratelimit

import (
	"net"
	"net/http"
	"strings"
)

// KeyFunc extracts a unique rate-limiting identifier from an incoming HTTP request.
// Universal interface: can extract IP, UserID, API Key, Header, or any custom attribute.
type KeyFunc func(r *http.Request) string

// KeyByIP extracts the client's real IP address from proxy headers or remote address.
// Checks X-Forwarded-For, X-Real-IP, and strips port numbers cleanly.
func KeyByIP(r *http.Request) string {
	// 1. Check X-Forwarded-For header (comma-separated list of proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return cleanIP(ip)
			}
		}
	}

	// 2. Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" {
			return cleanIP(ip)
		}
	}

	// 3. Fallback to r.RemoteAddr
	return cleanIP(r.RemoteAddr)
}

// KeyByHeader creates a KeyFunc that extracts value from a specified HTTP header.
// E.g. KeyByHeader("X-API-Key") or KeyByHeader("Authorization")
func KeyByHeader(headerName string) KeyFunc {
	return func(r *http.Request) string {
		val := strings.TrimSpace(r.Header.Get(headerName))
		if val == "" {
			return KeyByIP(r) // Fallback to IP if header is missing
		}
		return headerName + ":" + val
	}
}

// KeyByPathAndIP creates a KeyFunc that rate-limits by combination of Request Path + Client IP.
// E.g. Prevents brute-forcing /login specifically while allowing normal browsing elsewhere.
func KeyByPathAndIP(r *http.Request) string {
	return r.URL.Path + ":" + KeyByIP(r)
}

// KeyByComposite combines multiple KeyFuncs with a delimiter.
func KeyByComposite(funcs ...KeyFunc) KeyFunc {
	return func(r *http.Request) string {
		var parts []string
		for _, fn := range funcs {
			if fn != nil {
				if key := fn(r); key != "" {
					parts = append(parts, key)
				}
			}
		}
		if len(parts) == 0 {
			return KeyByIP(r)
		}
		return strings.Join(parts, ":")
	}
}

// cleanIP strips port numbers from IP strings (e.g. "192.168.1.1:54321" -> "192.168.1.1").
func cleanIP(hostPort string) string {
	host, _, err := net.SplitHostPort(hostPort)
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(hostPort)
}
