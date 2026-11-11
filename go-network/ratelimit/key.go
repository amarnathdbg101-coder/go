package ratelimit

import (
	"net"
	"net/http"
	"strings"
)

type KeyFunc func(r *http.Request) string

func KeyByIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return cleanIP(ip)
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" {
			return cleanIP(ip)
		}
	}

	return cleanIP(r.RemoteAddr)
}

func KeyByHeader(headerName string) KeyFunc {
	return func(r *http.Request) string {
		val := strings.TrimSpace(r.Header.Get(headerName))
		if val == "" {
			return KeyByIP(r)
		}
		return headerName + ":" + val
	}
}

func KeyByPathAndIP(r *http.Request) string {
	return r.URL.Path + ":" + KeyByIP(r)
}

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

func cleanIP(hostPort string) string {
	host, _, err := net.SplitHostPort(hostPort)
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(hostPort)
}
