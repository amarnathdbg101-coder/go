package secguard

import "errors"

var (
	// ErrMaliciousPattern indicates that an in-memory WAF pattern (SQLi, Traversal, XSS) was matched.
	ErrMaliciousPattern = errors.New("secguard: malicious payload pattern detected")

	// ErrExploitProbe indicates access attempted on known exploit probe paths (e.g. /.env, /wp-admin).
	ErrExploitProbe = errors.New("secguard: exploit probe path detected")

	// ErrScannerDetected indicates user agent matches known automated security scanning tools.
	ErrScannerDetected = errors.New("secguard: automated vulnerability scanner detected")

	// ErrIPQuarantined indicates the client IP is currently held in security quarantine.
	ErrIPQuarantined = errors.New("secguard: client IP is quarantined")

	// ErrHoneypotTriggered indicates a bot submitted data into a hidden honeypot trap field.
	ErrHoneypotTriggered = errors.New("secguard: honeypot bot trap triggered")
)
