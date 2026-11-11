package secguard

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig configures cross-origin resource sharing headers.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int // Seconds for Access-Control-Max-Age
}

// DefaultCORSConfig provides standard web and mobile client defaults.
var DefaultCORSConfig = CORSConfig{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "Origin"},
	ExposedHeaders:   []string{"Link", "X-Total-Count", "X-Request-ID"},
	AllowCredentials: true,
	MaxAge:           86400, // 24 hours
}

type wildcardRule struct {
	scheme string
	suffix string
}

// CORSMiddleware holds pre-computed header values for zero-allocation request handling.
type CORSMiddleware struct {
	cfg            CORSConfig
	allowAll       bool
	allowedOrigins map[string]bool
	wildcardRules  []wildcardRule
	methodsHeader  string
	headersHeader  string
	exposedHeader  string
	maxAgeHeader   string
}

// NewCORS initializes zero-allocation CORS middleware.
func NewCORS(cfg ...CORSConfig) *CORSMiddleware {
	c := DefaultCORSConfig
	if len(cfg) > 0 {
		c = cfg[0]
	}

	cors := &CORSMiddleware{
		cfg:            c,
		allowedOrigins: make(map[string]bool),
		methodsHeader:  strings.Join(c.AllowedMethods, ", "),
		headersHeader:  strings.Join(c.AllowedHeaders, ", "),
		exposedHeader:  strings.Join(c.ExposedHeaders, ", "),
		maxAgeHeader:   strconv.Itoa(c.MaxAge),
	}

	for _, o := range c.AllowedOrigins {
		trimmed := strings.TrimSpace(o)
		if trimmed == "*" {
			cors.allowAll = true
			break
		}
		if strings.HasPrefix(trimmed, "*.") || strings.Contains(trimmed, "://*.") {
			rule := wildcardRule{}
			if idx := strings.Index(trimmed, "://*."); idx != -1 {
				rule.scheme = trimmed[:idx+3]
				rule.suffix = trimmed[idx+5:]
			} else {
				rule.suffix = trimmed[1:]
			}
			cors.wildcardRules = append(cors.wildcardRules, rule)
		} else {
			cors.allowedOrigins[trimmed] = true
		}
	}

	return cors
}

// Handler wraps an http.Handler with CORS enforcement.
func (c *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !c.isOriginAllowed(origin) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Set("Vary", "Origin")

		if c.allowAll && !c.cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if c.cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if c.exposedHeader != "" {
			w.Header().Set("Access-Control-Expose-Headers", c.exposedHeader)
		}

		// Preflight request
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			if c.methodsHeader != "" {
				w.Header().Set("Access-Control-Allow-Methods", c.methodsHeader)
			}
			if c.headersHeader != "" {
				w.Header().Set("Access-Control-Allow-Headers", c.headersHeader)
			}
			if c.maxAgeHeader != "" {
				w.Header().Set("Access-Control-Max-Age", c.maxAgeHeader)
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c *CORSMiddleware) isOriginAllowed(origin string) bool {
	if c.allowAll {
		return true
	}
	if c.allowedOrigins[origin] {
		return true
	}
	for _, rule := range c.wildcardRules {
		if rule.scheme != "" && !strings.HasPrefix(origin, rule.scheme) {
			continue
		}
		domain := origin
		if idx := strings.Index(origin, "://"); idx != -1 {
			domain = origin[idx+3:]
		}
		if strings.HasSuffix(domain, rule.suffix) {
			return true
		}
	}
	return false
}
