package secguard

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	sqliRegex      = regexp.MustCompile(`(?i)(union\s+(all\s+)?select|select\s+[\w\*\s,]+\s+from|insert\s+into|delete\s+from|drop\s+(table|database)|alter\s+table|;\s*--|--\s*$|/\*.*\*/|or\s+['"\d\w]+=['"\d\w]+|benchmark\s*\(|sleep\s*\()`)
	traversalRegex = regexp.MustCompile(`(?i)(\.\.[/\\]\.\.[/\\]|/etc/passwd|/proc/self|windows[/\\]system32)`)
	xssRegex       = regexp.MustCompile(`(?i)(<script|javascript:|onerror\s*=|onload\s*=|document\.cookie|<iframe|<svg)`)
)

var knownScanners = []string{
	"sqlmap", "nikto", "masscan", "nmap", "zgrab",
	"gobuster", "dirbuster", "wpscan", "openvas", "nessus",
	"acunetix", "havij", "netsparker", "arachni",
}

var exploitProbes = []string{
	"/.env", "/.git", "/wp-admin", "/wp-login.php",
	"/xmlrpc.php", "/phpmyadmin", "/pma", "/.aws",
	"/actuator", "/solr", "/cgi-bin", "/shell.php",
	"/dump.sql", "/backup.sql", "/config.json",
}

// Engine encapsulates security inspect, WAF, and jail coordination.
type Engine struct {
	cfg *Config
}

// NewEngine creates a new Engine with options.
func NewEngine(opts ...Option) *Engine {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return &Engine{cfg: cfg}
}

// ExtractIP extracts real client IP from standard proxy headers or remote address.
func ExtractIP(r *http.Request) string {
	if cf := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cf != "" {
		return cf
	}
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		first := strings.TrimSpace(parts[0])
		if first != "" {
			return first
		}
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

// InjectSecurityHeaders writes OWASP headers to response writer.
func (e *Engine) InjectSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), payment=()")

	isTLS := r.TLS != nil ||
		strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") ||
		strings.Contains(r.Header.Get("CF-Visitor"), "https")

	if isTLS {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	}
}

func hasSuspiciousQuery(q string) bool {
	for i := 0; i < len(q); i++ {
		c := q[i]
		if c == '\'' || c == '"' || c == ';' || c == '<' || c == '>' ||
			c == '\\' || c == '(' || c == ')' || c == '*' || c == '-' {
			return true
		}
	}
	return false
}

func hasSuspiciousPath(p string) bool {
	for i := 0; i < len(p); i++ {
		c := p[i]
		if c == '\'' || c == '"' || c == ';' || c == '<' || c == '>' ||
			c == '\\' || c == '(' || c == ')' || c == '.' {
			return true
		}
	}
	return false
}

// InspectWAF checks path and raw query against SQLi, XSS, and Traversal patterns.
// Features a sub-microsecond fast-path filter that skips heavy regex evaluation on clean alphanumeric requests.
func (e *Engine) InspectWAF(path, query string) (bool, string) {
	// Fast-path: If neither path nor query contains suspicious injection punctuation, skip heavy regex engines.
	if !hasSuspiciousPath(path) && !hasSuspiciousQuery(query) {
		return false, ""
	}

	decodedQuery, err := url.QueryUnescape(query)
	if err != nil {
		decodedQuery = query
	}
	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		decodedPath = path
	}

	target := decodedPath + " " + decodedQuery
	if sqliRegex.MatchString(target) {
		return true, "SQL Injection attempt"
	}
	if traversalRegex.MatchString(target) {
		return true, "Path Traversal attempt"
	}
	if xssRegex.MatchString(target) {
		return true, "Cross-Site Scripting (XSS) attempt"
	}
	return false, ""
}

// IsExploitProbe checks if target path is a known vulnerability probe.
func (e *Engine) IsExploitProbe(path string) bool {
	lower := strings.ToLower(path)
	for _, probe := range exploitProbes {
		if strings.Contains(lower, probe) {
			return true
		}
	}
	return false
}

// IsMaliciousScanner checks user agent string.
func (e *Engine) IsMaliciousScanner(ua string) bool {
	if ua == "" {
		return false
	}
	lower := strings.ToLower(ua)
	for _, s := range knownScanners {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// IsExemptPath verifies whether current URL is in whitelist.
func (e *Engine) IsExemptPath(path string) bool {
	for _, ex := range e.cfg.ExemptPaths {
		if path == ex || strings.HasPrefix(path, ex) {
			return true
		}
	}
	return false
}

// JailIP places an offending IP into quarantine.
func (e *Engine) JailIP(ctx context.Context, ip string, reason string) {
	_ = e.cfg.JailStore.Jail(ctx, ip, e.cfg.JailDuration, reason)
	if e.cfg.Hooks.OnIPJailed != nil {
		e.cfg.Hooks.OnIPJailed(ip, e.cfg.JailDuration, reason)
	}
}

// Middleware returns the HTTP security guard middleware.
func (e *Engine) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ip := ExtractIP(r)

		// 1. Inject OWASP Security Headers
		if e.cfg.EnableSecurityHeaders {
			e.InjectSecurityHeaders(w, r)
		}

		// 2. Reject dangerous non-standard HTTP methods
		if r.Method == "TRACE" || r.Method == "CONNECT" || r.Method == "TRACK" {
			http.Error(w, "HTTP method not permitted", http.StatusMethodNotAllowed)
			return
		}

		// 3. Skip checks on exempt health paths
		if e.IsExemptPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// 4. Quarantine Jail Check
		jailed, reason, retryAfter, _ := e.cfg.JailStore.IsJailed(ctx, ip)
		if jailed {
			w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
			http.Error(w, "Access suspended: "+reason, http.StatusForbidden)
			return
		}

		// 5. In-Memory WAF inspection
		if e.cfg.EnableWAF {
			if malicious, rule := e.InspectWAF(r.URL.Path, r.URL.RawQuery); malicious {
				if e.cfg.Hooks.OnWAFViolation != nil {
					e.cfg.Hooks.OnWAFViolation(ip, rule, r.URL.Path)
				}
				e.JailIP(ctx, ip, "WAF violation: "+rule)
				http.Error(w, "Malicious request pattern detected", http.StatusForbidden)
				return
			}
		}

		// 6. Exploit probe detection
		if e.cfg.EnableExploitProbeBlock && e.IsExploitProbe(r.URL.Path) {
			if e.cfg.Hooks.OnExploitProbe != nil {
				e.cfg.Hooks.OnExploitProbe(ip, r.URL.Path)
			}
			e.JailIP(ctx, ip, "exploit probe: "+r.URL.Path)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// 7. Malicious scanner UA detection
		if e.cfg.EnableScannerBlock && e.IsMaliciousScanner(r.UserAgent()) {
			if e.cfg.Hooks.OnScannerBlocked != nil {
				e.cfg.Hooks.OnScannerBlocked(ip, r.UserAgent())
			}
			e.JailIP(ctx, ip, "scanner detected: "+r.UserAgent())
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// 8. Body size guard
		if e.cfg.MaxBodyBytes > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, e.cfg.MaxBodyBytes)
		}

		next.ServeHTTP(w, r)
	})
}

// Close cleans up underlying jail store.
func (e *Engine) Close() error {
	if e.cfg != nil && e.cfg.JailStore != nil {
		return e.cfg.JailStore.Close()
	}
	return nil
}
