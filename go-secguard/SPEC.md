# SPEC: go-secguard

## 1. Scope & Purpose
`go-secguard` provides zero-dependency OWASP production security headers, in-memory WAF pattern matching (SQLi, Directory Traversal, XSS), automated scanner blocking (sqlmap, nikto), exploit path trapping (/.env, /.git), and an auto-expiring IP quarantine jail.

## 2. Public API Contract
- `type JailStorage interface`: Contract for IP quarantine state.
- `type MemoryJail struct`: In-memory bounded cache with automatic background janitor.
- `type Engine struct`: Request inspector and header injector.
- `func Middleware(opts ...Option) func(http.Handler) http.Handler`: Standard net/http middleware.
- `func ValidateHoneypot(honeypotVal string, ip string, jail JailStorage, duration time.Duration) bool`: Honeypot bot detector.

## 3. Concurrency & Memory Guarantees
- Pre-compiled regexes (`regexp.MustCompile`) initialized once at package startup to avoid runtime compilation overhead.
- Memory-bounded IP quarantine: capped at `maxEntries` (default 10,000) to protect against memory starvation attacks.
- Graceful shutdown via `Close()`.

## 4. How to Extend
To share quarantine status across multiple Kubernetes pods or API instances, implement `JailStorage` using Redis via `secguard.WithJailStore(redisStore)`.
