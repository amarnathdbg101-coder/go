package secguard

import "time"

// Option configures SecGuard settings.
type Option func(*Config)

// Hooks defines zero-dependency security telemetry and incident reporting callbacks.
type Hooks struct {
	// OnWAFViolation is invoked when a SQLi, XSS, or Directory Traversal attempt is intercepted.
	OnWAFViolation func(ip string, rule string, path string)

	// OnScannerBlocked is invoked when an automated security tool (e.g. sqlmap, nikto) is detected.
	OnScannerBlocked func(ip string, userAgent string)

	// OnExploitProbe is invoked when an attacker attempts to probe known vulnerability files (e.g. /.env, /.git).
	OnExploitProbe func(ip string, path string)

	// OnIPJailed is invoked when an offending client IP is quarantined.
	OnIPJailed func(ip string, duration time.Duration, reason string)
}

// Config defines active defenses and thresholds.
type Config struct {
	EnableSecurityHeaders   bool
	EnableWAF               bool
	EnableScannerBlock      bool
	EnableExploitProbeBlock bool
	JailDuration            time.Duration
	MaxBodyBytes            int64
	ExemptPaths             []string
	JailStore               JailStorage
	Hooks                   Hooks
}

// DefaultConfig provides recommended security defaults.
func DefaultConfig() *Config {
	return &Config{
		EnableSecurityHeaders:   true,
		EnableWAF:               true,
		EnableScannerBlock:      true,
		EnableExploitProbeBlock: true,
		JailDuration:            30 * time.Minute,
		MaxBodyBytes:            10 << 20, // 10MB
		ExemptPaths:             []string{"/health", "/livez", "/readyz"},
		JailStore:               NewMemoryJail(10000, 1*time.Minute),
	}
}

// WithSecurityHeaders enables or disables HTTP security headers injection.
func WithSecurityHeaders(enable bool) Option {
	return func(c *Config) {
		c.EnableSecurityHeaders = enable
	}
}

// WithWAF enables or disables query/path pattern inspection.
func WithWAF(enable bool) Option {
	return func(c *Config) {
		c.EnableWAF = enable
	}
}

// WithJailDuration sets duration for IP quarantine.
func WithJailDuration(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.JailDuration = d
		}
	}
}

// WithMaxBodyBytes sets maximum allowed payload size.
func WithMaxBodyBytes(limit int64) Option {
	return func(c *Config) {
		if limit > 0 {
			c.MaxBodyBytes = limit
		}
	}
}

// WithExemptPaths adds endpoints that bypass quarantine and WAF checks.
func WithExemptPaths(paths ...string) Option {
	return func(c *Config) {
		c.ExemptPaths = append(c.ExemptPaths, paths...)
	}
}

// WithJailStore sets a custom quarantine storage driver.
func WithJailStore(store JailStorage) Option {
	return func(c *Config) {
		if store != nil {
			c.JailStore = store
		}
	}
}

// WithHooks attaches incident reporting and telemetry callbacks.
func WithHooks(hooks Hooks) Option {
	return func(c *Config) {
		c.Hooks = hooks
	}
}

